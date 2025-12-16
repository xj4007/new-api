HTML格式要求总结
### ✅ 可以使用的特性

1. **布局方式**
   - ✅ table布局（最稳定，推荐）
   - ✅ 简单的div + inline样式
   - ✅ 基础的margin、padding
   - ❌ 复杂的flexbox（display: flex可能不稳定）
   - ❌ grid布局

2. **CSS样式**
   - ✅ 在<style>标签中定义简单的class
   - ✅ 内联样式（style="..."）
   - ✅ 基础属性：background, color, padding, margin, border-radius
   - ✅ 简单渐变：linear-gradient
   - ✅ 基础动画：@keyframes + animation
   - ❌ backdrop-filter（毛玻璃效果）
   - ❌ 复杂的transform组合
   - ❌ clip-path等高级特性

3. **动画**
   - ✅ 简单的@keyframes动画（opacity, transform单独使用）
   - ✅ transition过渡效果
   - ❌ 过于复杂的动画组合

4. **图片**
   - ✅ <img>标签
   - ✅ 基础样式（width, height, border-radius）

5. **文本**
   - ✅ 基础标签：h1-h6, p, span, strong
   - ✅ font-size, color, font-weight
   - ✅ text-align

### 💡 最佳实践

1. **优先使用table布局**进行多列排版
2. **CSS class保持简单**，复杂样式用内联
3. **动画要简洁**，单一属性变化
4. **避免深层嵌套**div结构
5. **测试原则**：如果某段代码直接显示，说明有不兼容的属性，需要简化



现版本的代码:
<style>
  @keyframes pulse {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.05); }
  }
  @keyframes glow {
    0%, 100% { box-shadow: 0 0 20px rgba(255,107,53,0.3), 0 0 40px rgba(255,107,53,0.2); }
    50% { box-shadow: 0 0 30px rgba(255,107,53,0.5), 0 0 60px rgba(255,107,53,0.3); }
  }
  @keyframes float {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(-8px); }
  }
  .logo-container {
    animation: float 3s ease-in-out infinite;
  }
  .logo-glow {
    animation: glow 2s ease-in-out infinite;
  }
  .orange-box {
    background: linear-gradient(135deg, #ff6b35, #ff8e53);
    padding: 50px;
    border-radius: 24px;
    color: white;
    margin-bottom: 40px;
  }
  .white-card {
    background: white;
    padding: 30px;
    border-radius: 20px;
    margin: 20px 10px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  }
  .gray-bg {
    background: #f5f5f7;
    padding: 50px 30px;
    border-radius: 24px;
    margin-bottom: 40px;
  }
</style>

<div style="max-width: 1200px; margin: 0 auto; padding: 30px 20px;">

<div style="text-align: center; margin-bottom: 30px;">
<div class="logo-container" style="display: inline-block; margin-bottom: 15px;">
<div class="logo-glow" style="background: linear-gradient(135deg, #ff6b35, #ff8e53); padding: 8px; border-radius: 20px; display: inline-block; position: relative;">
<img src="https://new-api.xhm.gd.cn/logo.png" alt="Logo" style="width: 60px; height: 60px; border-radius: 14px; display: block;" />
<div style="position: absolute; top: -3px; right: -3px; background: #34c759; width: 16px; height: 16px; border-radius: 8px; border: 2px solid white;"></div>
</div>
</div>
<h1 style="font-size: 2rem; margin: 10px 0 5px 0; color: #1d1d1f;">小红帽 <span style="color: #ff6b35;">AICoding</span></h1>
<p style="font-size: 1rem; color: #86868b;">让开发者使用更具性价比的方式开发</p>
</div>

<div class="orange-box" style="padding: 25px; margin-bottom: 20px;">
<h2 style="font-size: 1.5rem; text-align: center; margin: 0 0 20px 0;">💎 透明计费模式</h2>

<table width="100%" cellpadding="0" cellspacing="10">
<tr>
<td width="50%">
<div class="white-card" style="padding: 20px; margin: 5px;">
<div style="font-size: 1.8rem; margin-bottom: 8px;">💰</div>
<h3 style="font-size: 1.1rem; margin: 0 0 5px 0; color: #1d1d1f;">统一汇率</h3>
<p style="color: #86868b; margin: 0 0 12px 0; font-size: 0.85rem;">按量付费，告别包月浪费</p>
<div style="background: #fff5f0; color: #ff6b35; padding: 12px; border-radius: 8px; text-align: center; font-weight: bold; font-size: 1rem;">
1 人民币 = 1 美元额度
</div>
</div>
</td>
<td width="50%">
<div class="white-card" style="padding: 20px; margin: 5px;">
<div style="font-size: 1.8rem; margin-bottom: 8px;">🎯</div>
<h3 style="font-size: 1.1rem; margin: 0 0 5px 0; color: #1d1d1f;">渠道选择</h3>
<p style="color: #86868b; margin: 0 0 12px 0; font-size: 0.85rem;">三大渠道，总有一款适合你</p>
<div style="background: #f5f5f7; padding: 10px; border-radius: 8px; border: 1px solid #eee;">
<p style="margin: 0; color: #1d1d1f; font-size: 0.9rem;">查看下方渠道详情 <strong style="color: #34c759; float: right;">↓</strong></p>
</div>
</div>
</td>
</tr>
</table>

</div>

<div class="gray-bg" style="padding: 25px 20px; margin-bottom: 20px;">
<h3 style="font-size: 1.3rem; text-align: center; margin: 0 0 20px 0; color: #1d1d1f;">🚀 四大分组任你选</h3>

<table width="100%" cellpadding="0" cellspacing="8">
<tr>
<td width="25%">
<div style="background: white; padding: 15px; border-radius: 12px; border: 2px solid #ff6b35; text-align: center;">
<div style="font-size: 1.8rem; margin-bottom: 6px;">👑</div>
<h4 style="color: #ff6b35; margin: 0 0 6px 0; font-size: 1rem;">纯max号池</h4>
<p style="color: #666; margin: 0; line-height: 1.4; font-size: 0.8rem;">
<strong style="color: #ff6b35;">0.65倍率</strong><br>
<strong>200K 有缓存</strong><br>
额度有限尽量保障稳定QAQ<br>
<span style="background: #fff5f0; padding: 3px 8px; border-radius: 5px; display: inline-block; margin-top: 6px; font-size: 0.75rem;">💎 稳定至上</span>
</p>
</div>
</td>
<td width="25%">
<div style="background: white; padding: 15px; border-radius: 12px; border: 2px solid #34c759; text-align: center;">
<div style="font-size: 1.8rem; margin-bottom: 6px;">⭐</div>
<h4 style="color: #34c759; margin: 0 0 6px 0; font-size: 1rem;">多渠道号池 🔥</h4>
<strong style="color: #34c759;">0.45倍率</strong><br>
<p style="color: #666; margin: 0; line-height: 1.4; font-size: 0.8rem;">
<strong>200K 有缓存</strong><br>
<strong style="color: #34c759;">SLA 99.99%</strong><br>
性价比之王！<br>
<span style="background: #e8f5e9; padding: 3px 8px; border-radius: 5px; display: inline-block; margin-top: 6px; font-size: 0.75rem;">🎯 站长推荐</span>
</p>
</div>
</td>
<td width="25%">
<div style="background: white; padding: 15px; border-radius: 12px; border: 2px solid #9c27b0; text-align: center;">
<div style="font-size: 1.8rem; margin-bottom: 6px;">⚡</div>
<h4 style="color: #9c27b0; margin: 0 0 6px 0; font-size: 1rem;">逆向渠道</h4>
<p style="color: #666; margin: 0; line-height: 1.4; font-size: 0.8rem;">
<strong style="color: #9c27b0;">0.25倍率</strong><br>
<strong>128K</strong> <strong style="color:rgb(201, 34, 34);">无缓存</strong><br>
短上下文<strong style="color: #9c27b0;">无敌</strong>！<br>
<span style="background: #f3e5f5; padding: 3px 8px; border-radius: 5px; display: inline-block; margin-top: 6px; font-size: 0.75rem;">💰 超低价格</span>
</p>
</div>
</td>
<td width="25%">
<div style="background: white; padding: 15px; border-radius: 12px; border: 2px solid #2196F3; text-align: center;">
<div style="font-size: 1.8rem; margin-bottom: 6px;">🤖</div>
<h4 style="color: #2196F3; margin: 0 0 6px 0; font-size: 1rem;">Codex渠道</h4>
<p style="color: #666; margin: 0; line-height: 1.4; font-size: 0.8rem;">
<strong style="color: #2196F3;">0.3倍率</strong><br>
<strong>仅限Codex终端</strong><br>
<span style="background: #E3F2FD; padding: 3px 8px; border-radius: 5px; display: inline-block; margin-top: 6px; font-size: 0.75rem;">🔒 OPENAI</span>
</p>
</div>
</td>
</tr>
</table>

<div style="background: #fff5f0; padding: 12px; border-radius: 10px; border-left: 3px solid #ff6b35; margin-top: 12px;">
<h4 style="color: #ff6b35; margin: 0 0 6px 0; font-size: 0.9rem;">⚠️ 重要说明</h4>
<p style="color: #666; margin: 0; line-height: 1.4; font-size: 0.8rem;">
• <strong>Claude渠道</strong>（前三个）：仅限 Claude Code 终端或 VSCode 插件使用<br>
• <strong>Codex渠道</strong>：仅限 Codex 终端或 VSCode 插件使用<br>
• <strong>注意</strong>：禁止NSFW内容，站内有AI道德审查机制，请勿轻易尝试输入违规内容，违者封号退款处理
</p>
</div>

</div>

<table width="100%" cellpadding="0" cellspacing="10" style="margin-bottom: 20px;">
<tr>
<td width="50%">
<div class="white-card" style="text-align: center; padding: 15px; margin: 5px;">
<div style="font-size: 1.8rem; margin-bottom: 6px;">🎁</div>
<h4 style="margin: 0 0 5px 0; font-size: 0.95rem;">新人福利</h4>
<p style="margin: 0; color: #86868b; font-size: 0.8rem;">注册送 $3 | 邀请送 $5</p>
</div>
</td>
<td width="50%">
<div class="white-card" style="text-align: center; padding: 15px; margin: 5px;">
<div style="font-size: 1.8rem; margin-bottom: 6px;">🏢</div>
<h4 style="margin: 0 0 5px 0; font-size: 0.95rem;">企业服务</h4>
<p style="margin: 0; color: #86868b; font-size: 0.8rem;">支持开票</p>
</div>
</td>
</tr>
</table>

<div style="background: white; padding: 25px 20px; border-radius: 16px; text-align: center; box-shadow: 0 4px 16px rgba(255,107,53,0.1);">
<h3 style="font-size: 1.3rem; margin: 0 0 20px 0;">联系我们</h3>
<table width="100%" cellpadding="0" cellspacing="15">
<tr>
<td width="33%" style="text-align: center;">
<div style="background: linear-gradient(135deg, #ff6b35, #ff8e53); width: 48px; height: 48px; border-radius: 12px; margin: 0 auto 10px; line-height: 48px; font-size: 24px;">💬</div>
<p style="margin: 0 0 5px 0; color: #86868b; font-size: 0.85rem;">客服QQ</p>
<p style="margin: 0; font-weight: bold; color: #1d1d1f; font-size: 0.9rem;">暂无</p>
</td>
<td width="33%" style="text-align: center;">
<div style="background: linear-gradient(135deg, #ff6b35, #ff8e53); width: 48px; height: 48px; border-radius: 12px; margin: 0 auto 10px; line-height: 48px; font-size: 24px;">📱</div>
<p style="margin: 0 0 8px 0; color: #86868b; font-size: 0.85rem;">微信客服</p>
<div style="background: #f5f5f7; padding: 10px; border-radius: 12px; display: inline-block;">
<img src="https://new-api.xhm.gd.cn/wechat.jpg" alt="微信" style="width: 100px; height: 100px; border-radius: 8px; display: block;" />
<p style="margin: 6px 0 0 0; font-size: 0.75rem; color: #86868b;">扫码添加</p>
</div>
</td>
<td width="33%" style="text-align: center;">
<div style="background: linear-gradient(135deg, #ff6b35, #ff8e53); width: 48px; height: 48px; border-radius: 12px; margin: 0 auto 10px; line-height: 48px; font-size: 24px;">👥</div>
<p style="margin: 0 0 8px 0; color: #86868b; font-size: 0.85rem;">技术交流群</p>
<div style="background: #f5f5f7; padding: 10px; border-radius: 12px; display: inline-block;">
<img src="https://new-api.xhm.gd.cn/group.jpg" alt="群" style="width: 100px; height: 100px; border-radius: 8px; display: block;" />
<p style="margin: 6px 0 0 0; font-size: 0.75rem; color: #86868b;">扫码加入</p>
</div>
</td>
</tr>
</table>
<button style="margin-top: 20px; background: linear-gradient(135deg, #ff6b35, #ff8e53); color: white; border: none; padding: 12px 35px; border-radius: 100px; font-size: 0.95rem; font-weight: bold; cursor: pointer; box-shadow: 0 4px 16px rgba(255,107,53,0.3);">立即开始使用 →</button>
</div>

<div style="background: #1d1d1f; padding: 30px 25px; border-radius: 20px; margin-top: 25px;">
<h3 style="font-size: 1.3rem; text-align: center; margin: 0 0 8px 0; color: white;">📚 配置指南</h3>
<p style="text-align: center; margin: 0 0 20px 0; color: #86868b; font-size: 0.85rem;">详细文档：<a href="https://ai.feishu.cn/wiki/QsHwwkrKziq0rjkzId0cXqO8nU9" style="color: #ff8e53;">https://ai.feishu.cn/wiki/QsHwwkrKziq0rjkzId0cXqO8nU9</a></p>

<div style="background: #2d2d2f; padding: 20px; border-radius: 12px; margin-bottom: 15px;">
<h4 style="color: #ff8e53; margin: 0 0 15px 0; font-size: 1rem;">🔧 Claude Code - 一次性配置</h4>

<p style="color: #86868b; margin: 0 0 8px 0; font-size: 0.85rem;"><strong style="color: #34c759;">Linux / macOS:</strong></p>
<div style="background: #1a1a1c; padding: 12px 15px; border-radius: 8px; margin-bottom: 12px; font-family: 'Courier New', monospace; font-size: 0.8rem; color: #e0e0e0; overflow-x: auto;">
<code style="color: #ff8e53;">export</code> ANTHROPIC_BASE_URL=https://new-api.xhm.gd.cn<br>
<code style="color: #ff8e53;">export</code> ANTHROPIC_AUTH_TOKEN=您的 API Key
</div>

<p style="color: #86868b; margin: 0 0 8px 0; font-size: 0.85rem;"><strong style="color: #2196F3;">Windows (PowerShell):</strong></p>
<div style="background: #1a1a1c; padding: 12px 15px; border-radius: 8px; font-family: 'Courier New', monospace; font-size: 0.8rem; color: #e0e0e0; overflow-x: auto;">
<code style="color: #ff8e53;">$env:</code>ANTHROPIC_BASE_URL=<code style="color: #34c759;">"https://new-api.xhm.gd.cn"</code><br>
<code style="color: #ff8e53;">$env:</code>ANTHROPIC_AUTH_TOKEN=<code style="color: #34c759;">"您的 API Key"</code>
</div>
</div>

<div style="background: #2d2d2f; padding: 20px; border-radius: 12px; margin-bottom: 15px;">
<h4 style="color: #ff8e53; margin: 0 0 15px 0; font-size: 1rem;">💾 Claude Code - 持久化配置</h4>
<p style="color: #86868b; margin: 0 0 8px 0; font-size: 0.85rem;">修改 <code style="background: #1a1a1c; padding: 2px 6px; border-radius: 4px; color: #e0e0e0;">~/.claude/settings.json</code>：</p>
<div style="background: #1a1a1c; padding: 12px 15px; border-radius: 8px; font-family: 'Courier New', monospace; font-size: 0.8rem; color: #e0e0e0; overflow-x: auto;">
{<br>
&nbsp;&nbsp;&nbsp;&nbsp;<code style="color: #34c759;">"env"</code>: {<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<code style="color: #34c759;">"ANTHROPIC_BASE_URL"</code>: <code style="color: #ff8e53;">"https://new-api.xhm.gd.cn"</code>,<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<code style="color: #34c759;">"ANTHROPIC_AUTH_TOKEN"</code>: <code style="color: #ff8e53;">"您的 API Key"</code><br>
&nbsp;&nbsp;&nbsp;&nbsp;}<br>
}
</div>
</div>

<div style="background: #2d2d2f; padding: 20px; border-radius: 12px;">
<h4 style="color: #2196F3; margin: 0 0 15px 0; font-size: 1rem;">🤖 Codex - 配置指南</h4>

<table width="100%" cellpadding="0" cellspacing="10">
<tr>
<td width="50%" style="vertical-align: top;">
<p style="color: #86868b; margin: 0 0 8px 0; font-size: 0.85rem;"><strong style="color: #2196F3;">Windows 路径:</strong></p>
<p style="color: #e0e0e0; margin: 0 0 10px 0; font-size: 0.75rem;">
<code style="background: #1a1a1c; padding: 2px 6px; border-radius: 4px;">%USERPROFILE%\.codex\config.toml</code><br>
<code style="background: #1a1a1c; padding: 2px 6px; border-radius: 4px;">%USERPROFILE%\.codex\auth.json</code>
</p>
</td>
<td width="50%" style="vertical-align: top;">
<p style="color: #86868b; margin: 0 0 8px 0; font-size: 0.85rem;"><strong style="color: #34c759;">Linux / macOS 路径:</strong></p>
<p style="color: #e0e0e0; margin: 0 0 10px 0; font-size: 0.75rem;">
<code style="background: #1a1a1c; padding: 2px 6px; border-radius: 4px;">~/.codex/config.toml</code><br>
<code style="background: #1a1a1c; padding: 2px 6px; border-radius: 4px;">~/.codex/auth.json</code>
</p>
</td>
</tr>
</table>

<p style="color: #86868b; margin: 15px 0 8px 0; font-size: 0.85rem;"><strong style="color: #ff8e53;">config.toml 内容:</strong></p>
<div style="background: #1a1a1c; padding: 12px 15px; border-radius: 8px; margin-bottom: 12px; font-family: 'Courier New', monospace; font-size: 0.8rem; color: #e0e0e0; overflow-x: auto;">
<code style="color: #9c27b0;">model_provider</code> = <code style="color: #34c759;">"xhm"</code><br>
<code style="color: #9c27b0;">model</code> = <code style="color: #34c759;">"gpt-5-codex"</code><br>
<code style="color: #9c27b0;">model_reasoning_effort</code> = <code style="color: #34c759;">"high"</code><br>
<code style="color: #9c27b0;">disable_response_storage</code> = <code style="color: #ff8e53;">true</code><br>
<br>
<code style="color: #86868b;">[model_providers.xhm]</code><br>
<code style="color: #9c27b0;">name</code> = <code style="color: #34c759;">"xhm"</code><br>
<code style="color: #9c27b0;">base_url</code> = <code style="color: #34c759;">"https://new-api.xhm.gd.cn/v1"</code><br>
<code style="color: #9c27b0;">wire_api</code> = <code style="color: #34c759;">"responses"</code><br>
<code style="color: #9c27b0;">env_key</code> = <code style="color: #34c759;">"custom"</code><br>
<code style="color: #9c27b0;">requires_openai_auth</code> = <code style="color: #ff8e53;">true</code>
</div>

<p style="color: #86868b; margin: 0 0 8px 0; font-size: 0.85rem;"><strong style="color: #ff8e53;">auth.json 内容:</strong></p>
<div style="background: #1a1a1c; padding: 12px 15px; border-radius: 8px; font-family: 'Courier New', monospace; font-size: 0.8rem; color: #e0e0e0; overflow-x: auto;">
{<br>
&nbsp;&nbsp;&nbsp;&nbsp;<code style="color: #34c759;">"OPENAI_API_KEY"</code>: <code style="color: #ff8e53;">"您的 API Key"</code><br>
}
</div>
</div>

</div>

</div>