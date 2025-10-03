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

<div style="max-width: 1200px; margin: 0 auto; padding: 60px 20px;">

<div style="text-align: center; margin-bottom: 60px;">
<div class="logo-container" style="display: inline-block; margin-bottom: 24px;">
<div class="logo-glow" style="background: linear-gradient(135deg, #ff6b35, #ff8e53); padding: 10px; border-radius: 28px; display: inline-block; position: relative;">
<img src="https://new-api.xhm.gd.cn/logo.png" alt="Logo" style="width: 88px; height: 88px; border-radius: 20px; display: block;" />
<div style="position: absolute; top: -5px; right: -5px; background: #34c759; width: 20px; height: 20px; border-radius: 10px; border: 3px solid white;"></div>
</div>
</div>
<h1 style="font-size: 3rem; margin: 20px 0 10px 0; color: #1d1d1f;">小红帽 <span style="color: #ff6b35;">AICoding</span></h1>
<p style="font-size: 1.3rem; color: #86868b;">让开发者使用更具性价比的方式开发</p>
</div>

<div class="orange-box">
<h2 style="font-size: 2.5rem; text-align: center; margin: 0 0 40px 0;">💎 透明计费模式</h2>

<table width="100%" cellpadding="0" cellspacing="20">
<tr>
<td width="50%">
<div class="white-card">
<div style="font-size: 2.5rem; margin-bottom: 15px;">💰</div>
<h3 style="font-size: 1.3rem; margin: 0 0 8px 0; color: #1d1d1f;">统一汇率</h3>
<p style="color: #86868b; margin: 0 0 20px 0;">按量付费，告别包月浪费</p>
<div style="background: #fff5f0; color: #ff6b35; padding: 20px; border-radius: 12px; text-align: center; font-weight: bold; font-size: 1.2rem;">
1 人民币 = 1 美元额度
</div>
</div>
</td>
<td width="50%">
<div class="white-card">
<div style="font-size: 2.5rem; margin-bottom: 15px;">🎯</div>
<h3 style="font-size: 1.3rem; margin: 0 0 8px 0; color: #1d1d1f;">超值倍率</h3>
<p style="color: #86868b; margin: 0 0 20px 0;">两重优惠，省上加省</p>
<div style="background: white; padding: 15px; border-radius: 12px; border: 1px solid #eee;">
<p style="margin: 0 0 10px 0; padding-bottom: 10px; border-bottom: 1px solid #f5f5f5;">基础倍率 <strong style="color: #34c759; float: right;">0.5倍</strong></p>
<p style="margin: 0;">缓存倍率 <strong style="color: #34c759; float: right;">0.1倍</strong></p>
</div>
</div>
</td>
</tr>
</table>

</div>

<div class="gray-bg">
<h3 style="font-size: 2rem; text-align: center; margin: 0 0 40px 0; color: #1d1d1f;">📊 实际案例对比</h3>

<div style="background: white; padding: 40px; border-radius: 20px; margin-bottom: 20px;">
<table width="100%" cellpadding="20">
<tr>
<td width="40%" style="text-align: center;">
<p style="color: #86868b; margin: 0 0 10px 0;">官方定价</p>
<p style="font-size: 2.5rem; font-weight: bold; color: #1d1d1f; margin: 0;">$5 <span style="font-size: 1rem; color: #86868b;">美元</span></p>
</td>
<td width="20%" style="text-align: center; font-size: 2rem; color: #ff6b35;">→</td>
<td width="40%" style="text-align: center;">
<p style="color: #86868b; margin: 0 0 10px 0;">您实际支付</p>
<p style="font-size: 2.5rem; font-weight: bold; color: #34c759; margin: 0;">￥1 <span style="font-size: 1rem; color: #86868b;">人民币</span></p>
</td>
</tr>
</table>

<div style="background: #fff5f0; padding: 25px; border-radius: 16px; border-left: 4px solid #ff6b35; margin-top: 20px;">
<h4 style="color: #ff6b35; margin: 0 0 10px 0;">💡 智能缓存让您更省钱</h4>
<p style="color: #666; margin: 0; line-height: 1.6;">当您重复使用相同的提示词时，系统自动启用缓存机制，缓存内容仅收取10%费用。例如：同样的代码分析任务，第二次调用可能只需 ￥0.1-0.3 即可完成。</p>
</div>
</div>

<div style="background: white; padding: 30px; border-radius: 20px; border: 2px solid #ff6b35;">
<h4 style="margin: 0 0 20px 0; color: #1d1d1f;">🔍 费用计算原理</h4>
<p style="background: #ff6b35; color: white; padding: 8px 15px; border-radius: 8px; display: inline-block; margin: 0 0 8px 0; font-weight: bold;">步骤1</p>
<p style="color: #666; margin: 0 0 15px 20px;">官方原价：输入 $15/百万字符，输出 $75/百万字符</p>
<p style="background: #ff6b35; color: white; padding: 8px 15px; border-radius: 8px; display: inline-block; margin: 0 0 8px 0; font-weight: bold;">步骤2</p>
<p style="color: #666; margin: 0 0 15px 20px;">应用当前倍率 0.5：所有费用打5折 💸</p>
<p style="background: #ff6b35; color: white; padding: 8px 15px; border-radius: 8px; display: inline-block; margin: 0 0 8px 0; font-weight: bold;">步骤3</p>
<p style="color: #666; margin: 0 0 15px 20px;">缓存优化：已缓存内容仅收10%（48,595字符缓存仅需 $0.07）</p>
<p style="background: #34c759; color: white; padding: 8px 15px; border-radius: 8px; display: inline-block; margin: 0 0 8px 0; font-weight: bold;">结果</p>
<p style="color: #1d1d1f; margin: 0 0 0 20px;"><strong style="color: #34c759; font-size: 1.2rem;">最终仅需支付 ￥0.11</strong> <span style="color: #86868b;">（原价需 $0.38）</span></p>
</div>

</div>

<table width="100%" cellpadding="0" cellspacing="20" style="margin-bottom: 40px;">
<tr>
<td width="25%">
<div class="white-card" style="text-align: center;">
<div style="font-size: 2.5rem; margin-bottom: 12px;">⚡</div>
<h4 style="margin: 0 0 8px 0; font-size: 1.1rem;">极速响应</h4>
<p style="margin: 0; color: #86868b; font-size: 0.9rem;">毫秒级延迟<br>99.9% 可用性</p>
</div>
</td>
<td width="25%">
<div class="white-card" style="text-align: center;">
<div style="font-size: 2.5rem; margin-bottom: 12px;">🔒</div>
<h4 style="margin: 0 0 8px 0; font-size: 1.1rem;">企业级安全</h4>
<p style="margin: 0; color: #86868b; font-size: 0.9rem;">端到端加密<br>ISO认证</p>
</div>
</td>
<td width="25%">
<div class="white-card" style="text-align: center;">
<div style="font-size: 2.5rem; margin-bottom: 12px;">🎁</div>
<h4 style="margin: 0 0 8px 0; font-size: 1.1rem;">新人福利</h4>
<p style="margin: 0; color: #86868b; font-size: 0.9rem;">注册送 $10<br>立即开始</p>
</div>
</td>
<td width="25%">
<div class="white-card" style="text-align: center;">
<div style="font-size: 2.5rem; margin-bottom: 12px;">🏢</div>
<h4 style="margin: 0 0 8px 0; font-size: 1.1rem;">企业服务</h4>
<p style="margin: 0; color: #86868b; font-size: 0.9rem;">支持开票<br>专属客服</p>
</div>
</td>
</tr>
</table>

<div style="background: white; padding: 50px 30px; border-radius: 24px; text-align: center; box-shadow: 0 8px 32px rgba(255,107,53,0.15);">
<h3 style="font-size: 2rem; margin: 0 0 40px 0;">联系我们</h3>
<table width="100%" cellpadding="0" cellspacing="30">
<tr>
<td width="33%" style="text-align: center;">
<div style="background: linear-gradient(135deg, #ff6b35, #ff8e53); width: 64px; height: 64px; border-radius: 16px; margin: 0 auto 15px; line-height: 64px; font-size: 32px;">💬</div>
<p style="margin: 0 0 8px 0; color: #86868b; font-size: 0.9rem;">客服QQ</p>
<p style="margin: 0; font-weight: bold; color: #1d1d1f;">暂无</p>
</td>
<td width="33%" style="text-align: center;">
<div style="background: linear-gradient(135deg, #ff6b35, #ff8e53); width: 64px; height: 64px; border-radius: 16px; margin: 0 auto 15px; line-height: 64px; font-size: 32px;">📱</div>
<p style="margin: 0 0 12px 0; color: #86868b; font-size: 0.9rem;">微信客服</p>
<div style="background: #f5f5f7; padding: 15px; border-radius: 16px; display: inline-block;">
<img src="https://new-api.xhm.gd.cn/wechat.jpg" alt="微信" style="width: 150px; height: 150px; border-radius: 12px; display: block;" />
<p style="margin: 10px 0 0 0; font-size: 0.8rem; color: #86868b;">扫码添加客服</p>
</div>
</td>
<td width="33%" style="text-align: center;">
<div style="background: linear-gradient(135deg, #ff6b35, #ff8e53); width: 64px; height: 64px; border-radius: 16px; margin: 0 auto 15px; line-height: 64px; font-size: 32px;">👥</div>
<p style="margin: 0 0 12px 0; color: #86868b; font-size: 0.9rem;">技术交流群</p>
<div style="background: #f5f5f7; padding: 15px; border-radius: 16px; display: inline-block;">
<img src="https://new-api.xhm.gd.cn/group.jpg" alt="群" style="width: 150px; height: 150px; border-radius: 12px; display: block;" />
<p style="margin: 10px 0 0 0; font-size: 0.8rem; color: #86868b;">扫码加入群聊</p>
</div>
</td>
</tr>
</table>
<button style="margin-top: 40px; background: linear-gradient(135deg, #ff6b35, #ff8e53); color: white; border: none; padding: 18px 50px; border-radius: 100px; font-size: 1.1rem; font-weight: bold; cursor: pointer; box-shadow: 0 6px 24px rgba(255,107,53,0.3);">立即开始使用 →</button>
</div>

</div>