/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useCallback, useMemo, useState } from 'react';
import {
  Card,
  Input,
  Button,
  Table,
  Space,
  Typography,
  Banner,
  Spin,
  Empty,
} from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, timestamp2string } from '../../helpers';
import './TokenUsage.css';

const LOG_TYPE_LABELS = {
  1: '充值',
  2: '消费',
  3: '管理',
  4: '系统',
  5: '错误',
};

const DEFAULT_PAGE_SIZE = 10;
const TOKENS_PER_DOLLAR = 500000;

const formatCurrencyFromTokens = (value, digits = 3) => {
  const tokens = Number(value) || 0;
  const dollars = tokens / TOKENS_PER_DOLLAR;
  const fixed = dollars.toFixed(digits);
  const trimmed = fixed.replace(/\.0+$/, '').replace(/(\.\d*[1-9])0+$/, '$1');
  return `$${trimmed}`;
};

const parseTotal = (value) => {
  if (typeof value === 'number') {
    return value;
  }
  const parsed = Number(value);
  return Number.isNaN(parsed) ? 0 : parsed;
};

const TokenUsage = () => {
  const { t } = useTranslation();
  const [tokenInput, setTokenInput] = useState('');
  const [usageData, setUsageData] = useState(null);
  const [logs, setLogs] = useState([]);
  const [usageLoading, setUsageLoading] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [queried, setQueried] = useState(false);
  const [queriedToken, setQueriedToken] = useState('');
  const [logPage, setLogPage] = useState(1);
  const [logPageSize, setLogPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [logTotal, setLogTotal] = useState(0);
  const [usageError, setUsageError] = useState('');
  const [logsError, setLogsError] = useState('');

  const combinedError = useMemo(() => {
    return [usageError, logsError].filter(Boolean).join('；');
  }, [usageError, logsError]);

  const extractErrorMessage = useCallback((error) => {
    if (!error) return '';
    if (typeof error === 'string') {
      return error;
    }
    const responseMessage =
      error?.response?.data?.message || error?.response?.data?.error;
    if (responseMessage) {
      return responseMessage;
    }
    return error?.message || '';
  }, []);

  const fetchUsageInfo = useCallback(
    async (token) => {
      setUsageLoading(true);
      try {
        const response = await API.get('/api/usage/token', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
          skipErrorHandler: true,
        });
        const payload = response.data;
        if (payload?.code) {
          setUsageData(payload.data);
          return '';
        }
        setUsageData(null);
        return payload?.message || t('查询失败');
      } catch (error) {
        setUsageData(null);
        return extractErrorMessage(error) || t('查询失败');
      } finally {
        setUsageLoading(false);
      }
    },
    [extractErrorMessage, t],
  );

  const fetchLogsInfo = useCallback(
    async (token, page, pageSize) => {
      setLogsLoading(true);
      const normalizedPage = Number(page) > 0 ? Number(page) : 1;
      const normalizedSizeRaw = Number(pageSize);
      const normalizedSize =
        normalizedSizeRaw > 0 ? normalizedSizeRaw : DEFAULT_PAGE_SIZE;
      try {
        const response = await API.get('/api/log/token', {
          params: {
            key: token,
            p: normalizedPage,
            size: normalizedSize,
            order: 'desc',
          },
          skipErrorHandler: true,
        });
        const payload = response.data;
        if (payload?.success) {
          const list = Array.isArray(payload.data) ? payload.data : [];
          const sorted = [...list].sort(
            (a, b) => (b?.created_at ?? 0) - (a?.created_at ?? 0),
          );
          const offset = (normalizedPage - 1) * normalizedSize;
          setLogs(
            sorted.map((item, index) => ({
              ...item,
              key: `${item.id ?? 'row'}-${item.created_at ?? index}-${offset + index}`,
            })),
          );
          const totalVal =
            parseTotal(payload.total) ||
            parseTotal(payload.pagination?.total) ||
            sorted.length;
          setLogTotal(totalVal);
          setLogPage(normalizedPage);
          setLogPageSize(normalizedSize);
          return '';
        }
        setLogs([]);
        setLogTotal(0);
        return payload?.message || t('调用日志获取失败');
      } catch (error) {
        setLogs([]);
        setLogTotal(0);
        return extractErrorMessage(error) || t('调用日志获取失败');
      } finally {
        setLogsLoading(false);
      }
    },
    [extractErrorMessage, t],
  );

  const handleQuery = useCallback(async () => {
    const trimmedToken = tokenInput.trim();
    if (!trimmedToken) {
      setUsageError(t('请输入令牌'));
      setLogsError('');
      setUsageData(null);
      setLogs([]);
      setLogTotal(0);
      setQueried(false);
      setQueriedToken('');
      setLogPage(1);
      setLogPageSize(DEFAULT_PAGE_SIZE);
      return;
    }

    const pageSize = logPageSize || DEFAULT_PAGE_SIZE;
    setUsageError('');
    setLogsError('');
    setQueried(true);
    setQueriedToken(trimmedToken);
    const [usageErr, logsErr] = await Promise.all([
      fetchUsageInfo(trimmedToken),
      fetchLogsInfo(trimmedToken, 1, pageSize),
    ]);
    setUsageError(usageErr);
    setLogsError(logsErr);
  }, [
    tokenInput,
    logPageSize,
    fetchUsageInfo,
    fetchLogsInfo,
    t,
  ]);

  const handleLogsPagination = useCallback(
    async (page, pageSize) => {
      const nextPage = Number(page) || 1;
      const nextSize =
        Number(pageSize) || logPageSize || DEFAULT_PAGE_SIZE;
      if (!queriedToken) {
        setLogPage(nextPage);
        setLogPageSize(nextSize);
        return;
      }
      const errorMsg = await fetchLogsInfo(queriedToken, nextPage, nextSize);
      setLogsError(errorMsg);
    },
    [fetchLogsInfo, logPageSize, queriedToken],
  );

  const stats = useMemo(() => {
    const unlimited = usageData?.unlimited_quota;
    const totalGranted = usageData?.total_granted ?? 0;
    const totalUsed = usageData?.total_used ?? 0;
    const totalAvailable = usageData?.total_available ?? 0;
    const expiresAt = usageData?.expires_at ?? 0;

    const formatQuota = (value) => formatCurrencyFromTokens(value, 3);

    const expiryLabel =
      expiresAt && expiresAt > 0 ? timestamp2string(expiresAt) : t('未知');

    return [
      {
        label: t('令牌总额'),
        value: unlimited ? t('无限额度') : formatQuota(totalGranted),
      },
      {
        label: t('剩余额度'),
        value: unlimited ? t('无限额度') : formatQuota(totalAvailable),
      },
      {
        label: t('已用额度'),
        value: formatQuota(totalUsed),
      },
      {
        label: t('有效期至'),
        value: expiryLabel,
      },
    ];
  }, [usageData, t]);

  const logColumns = useMemo(
    () => [
      {
        title: t('时间'),
        dataIndex: 'created_at',
        width: 180,
        render: (value) => (value ? timestamp2string(value) : '-'),
      },
      {
        title: t('类型'),
        dataIndex: 'type',
        width: 80,
        render: (value) => t(LOG_TYPE_LABELS[value] || '未知'),
      },
      {
        title: t('模型'),
        dataIndex: 'model_name',
        render: (value) => value || '-',
      },
      {
        title: t('消耗额度'),
        dataIndex: 'quota',
        width: 120,
        render: (value) => formatCurrencyFromTokens(value, 3),
      },
      {
        title: t('IP'),
        dataIndex: 'ip',
        render: (value) => value || '-',
      },
    ],
    [t],
  );

  return (
    <div className='token-usage-page'>
      <div className='container'>
        <Card className='search-card' bordered={false}>
          <div className='search-section'>
            <Typography.Title heading={3} className='search-title'>
              {t('令牌用量查询')}
            </Typography.Title>
            <Typography.Text className='search-description'>
              {t('输入您的API令牌，查看详细的使用情况和调用记录')}
            </Typography.Text>
            <Space align='center' spacing='medium' wrap className='search-inputs'>
              <Input
                className='token-input'
                placeholder={t('请输入令牌')}
                value={tokenInput}
                onChange={setTokenInput}
                onEnterPress={handleQuery}
                prefix={<IconSearch />}
              />
              <Button
                className='search-button'
                type='primary'
                icon={<IconSearch />}
                onClick={handleQuery}
                loading={usageLoading || logsLoading}
              >
                {t('查询')}
              </Button>
            </Space>
            {queriedToken && (
              <div className='current-token'>
                <Typography.Text className='token-label'>
                  {t('当前令牌')}:
                </Typography.Text>
                <Typography.Text className='token-value' copyable>
                  {queriedToken}
                </Typography.Text>
              </div>
            )}
          </div>
        </Card>

        {combinedError && (
          <Banner
            type='warning'
            icon={null}
            className='error-banner'
            description={combinedError}
          />
        )}

        {queried && (
          <>
            <Card className='info-card' bordered={false}>
              <div className='card-header'>
                <Typography.Title heading={5} className='card-title'>
                  {t('令牌信息')}
                </Typography.Title>
              </div>
              <Spin spinning={usageLoading}>
                {usageData ? (
                  <div className='token-info'>
                    {usageData?.name && (
                      <div className='token-name-section'>
                        <Typography.Title heading={4} className='token-name'>
                          {usageData.name}
                        </Typography.Title>
                      </div>
                    )}
                    <div className='stats-grid'>
                      {stats.map((stat) => (
                        <div
                          key={stat.label}
                          className='stat-card'
                        >
                          <div className='stat-label'>
                            <Typography.Text type='tertiary'>
                              {stat.label}
                            </Typography.Text>
                          </div>
                          <div className='stat-value'>
                            <Typography.Title heading={4}>
                              {stat.value}
                            </Typography.Title>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className='empty-state'>
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      title={t('暂无令牌信息')}
                      description={t('查询的令牌不存在或无效')}
                    />
                  </div>
                )}
              </Spin>
            </Card>

            <Card className='logs-card' bordered={false}>
              <div className='card-header'>
                <Typography.Title heading={5} className='card-title'>
                  {t('调用日志')}
                </Typography.Title>
                {logTotal > 0 && (
                  <Typography.Text type='tertiary' className='log-count'>
                    共 {logTotal} 条记录
                  </Typography.Text>
                )}
              </div>
              <Spin spinning={logsLoading}>
                <Table
                  className='logs-table'
                  columns={logColumns}
                  dataSource={logs}
                  rowKey='key'
                  pagination={{
                    currentPage: logPage,
                    pageSize: logPageSize,
                    total: logTotal,
                    showSizeChanger: true,
                    showQuickJumper: true,
                    pageSizeOpts: ['10', '20', '50', '100'],
                    onChange: (page, size) => {
                      handleLogsPagination(page, size);
                    },
                    onShowSizeChange: (curr, size) => {
                      handleLogsPagination(1, size);
                    },
                  }}
                  empty={
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      title={t('暂无调用记录')}
                      description={t('该令牌暂无调用记录')}
                    />
                  }
                  scroll={{ x: 'max-content' }}
                />
              </Spin>
            </Card>
          </>
        )}
      </div>
    </div>
  );
};

export default TokenUsage;
