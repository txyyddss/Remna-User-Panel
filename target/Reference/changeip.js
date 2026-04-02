/**
 * Cloudflare Worker for Self-Service IP Replacement
 * 
 * Required Environment Variables (Secrets/KV):
 * - REQUESTS (KV Namespace Binding)
 * - TG_BOT_TOKEN (Secret)
 * - TG_CHAT_ID (Secret, e.g., -1003493995915)
 * - SALT (Secret/Env, e.g., "homo114514" for the swap token)
 */

const ALLOWED_SQUADS = [
    "04d22a2e-1979-47b9-946b-8dbea5398811",
    "5899ea60-974e-4794-9f71-ed73c2f8b24c",
    "88a969cd-e313-440c-8ebd-4a53d8a79c3b",
    "225a9b69-8b7d-4c70-ab57-13f547d96f54",
    "ba55bf05-1fc8-4b8b-b4b2-b6f20715ef03",
    "e68a5b66-f684-434b-9298-65e3c3867237"
];

const COOLDOWN_HOURS = 6;

export default {
    async fetch(request, env, ctx) {
        const url = new URL(request.url);
        const path = url.pathname;

        if (request.method === "GET" && path === "/") {
            return await renderHTML();
        }
        if (request.method === "POST" && path === "/submit") {
            return await handleSubmit(request, env);
        }
        if (request.method === "POST" && path === "/webhook") {
            return await handleWebhook(request, env);
        }
        if (request.method === "GET" && path === "/lookup") {
            return await handleLookup(request, env);
        }
        if (request.method === "POST" && path === "/swap") {
            return await handleSwap(request, env);
        }
        if (request.method === "POST" && path === "/api/add-request") {
            return await handleApiAddRequest(request, env);
        }

        return new Response("Not Found", { status: 404 });
    }
};

async function renderHTML() {
    const html = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>TX Hinet IP自助更换</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    :root {
      --bg-color: #050505;
      --card-bg: #111;
      --primary: #00ff9d;
      --secondary: #00d2ff;
      --text: #e0e0e0;
      --error: #ff3c3c;
    }
    body {
      background-color: var(--bg-color);
      color: var(--text);
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      margin: 0;
      background-image: 
        linear-gradient(rgba(0, 255, 157, 0.03) 1px, transparent 1px),
        linear-gradient(90deg, rgba(0, 255, 157, 0.03) 1px, transparent 1px);
      background-size: 30px 30px;
    }
    .container {
      background: rgba(17, 17, 17, 0.9);
      padding: 2rem;
      border-radius: 12px;
      border: 1px solid #333;
      box-shadow: 0 0 20px rgba(0, 255, 157, 0.1);
      width: 100%;
      max-width: 400px;
      backdrop-filter: blur(10px);
    }
    h1 {
      text-align: center;
      color: var(--primary);
      text-transform: uppercase;
      letter-spacing: 2px;
      margin-bottom: 2rem;
      text-shadow: 0 0 10px rgba(0, 255, 157, 0.5);
    }
    .input-group {
      margin-bottom: 1.5rem;
    }
    label {
      display: block;
      margin-bottom: 0.5rem;
      color: var(--secondary);
      font-size: 0.9rem;
    }
    input {
      width: 100%;
      padding: 10px;
      background: #000;
      border: 1px solid #333;
      border-radius: 4px;
      color: #fff;
      box-sizing: border-box;
      transition: border-color 0.3s;
    }
    input:focus {
      outline: none;
      border-color: var(--primary);
      box-shadow: 0 0 5px rgba(0, 255, 157, 0.3);
    }
    button {
      width: 100%;
      padding: 12px;
      background: linear-gradient(45deg, var(--primary), var(--secondary));
      border: none;
      border-radius: 4px;
      color: #000;
      font-weight: bold;
      cursor: pointer;
      text-transform: uppercase;
      letter-spacing: 1px;
      transition: transform 0.2s, box-shadow 0.2s;
    }
    button:hover {
      transform: translateY(-2px);
      box-shadow: 0 0 15px rgba(0, 255, 157, 0.4);
    }
    #message {
      margin-top: 1rem;
      text-align: center;
      min-height: 20px;
      font-size: 0.9rem;
    }
    .success { color: var(--primary); }
    .error { color: var(--error); }
  </style>
</head>
<body>
  <div class="container">
    <h1>IP自助更换</h1>
    <div class="input-group">
      <label>订阅链接</label>
      <input type="text" id="subLink" placeholder="https://sub.1391399.xyz/..." required>
    </div>
    <div class="input-group">
      <label>更换原因</label>
      <input type="text" id="reason" placeholder="为什么你想要更换IP?" required>
    </div>
    <button onclick="submitRequest()" id="btn">发起更换</button>
    <div id="message"></div>
  </div>

  <script>
    async function submitRequest() {
      const btn = document.getElementById('btn');
      const msg = document.getElementById('message');
      const link = document.getElementById('subLink').value;
      const reason = document.getElementById('reason').value;

      if (!link || !reason) {
        msg.textContent = "❌ 请填写所有字段";
        msg.className = "error";
        return;
      }

      btn.disabled = true;
      btn.textContent = "提交中...";
      msg.textContent = "";

      try {
        const res = await fetch('/submit', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ link, reason })
        });
        const data = await res.json();
        
        if (res.ok) {
          msg.textContent = "✅ 请求提交成功！请等待群内投票审核";
          msg.className = "success";
          document.getElementById('subLink').value = '';
          document.getElementById('reason').value = '';
        } else {
          // 显示具体的错误信息
          msg.innerHTML = "❌ " + (data.error || "未知错误");
          // 如果有Telegram消息链接，显示可点击的链接
          if (data.message_link) {
            msg.innerHTML += '<br><a href="' + data.message_link + '" target="_blank" style="color: var(--secondary); text-decoration: underline;">📨 查看当前请求</a>';
          }
          msg.className = "error";
        }
      } catch (e) {
        msg.textContent = "❌ 网络错误，请稍后重试";
        msg.className = "error";
      } finally {
        btn.disabled = false;
        btn.textContent = "发起更换";
      }
    }
  </script>
</body>
</html>
  `;
    return new Response(html, { headers: { "Content-Type": "text/html" } });
}

async function handleSubmit(request, env) {
    try {
        const { link, reason } = await request.json();

        const match = link.match(/sub\.1391399\.xyz\/([a-zA-Z0-9-]+)/);
        if (!match) {
            return new Response(JSON.stringify({ error: "Invalid subscription URL format" }), { status: 400 });
        }
        const shortUuid = match[1];

        const panelResp = await fetch(PANEL_API_BASE + shortUuid, {
            headers: { "Authorization": `Bearer ${AUTH_TOKEN}` }
        });

        if (!panelResp.ok) {
            return new Response(JSON.stringify({ error: "Failed to fetch user info" }), { status: 400 });
        }

        const panelData = await panelResp.json();
        const user = panelData.response;

        if (!user) {
            return new Response(JSON.stringify({ error: "User not found" }), { status: 404 });
        }

        if (user.status !== "ACTIVE") {
            return new Response(JSON.stringify({ error: "Subscription is not ACTIVE" }), { status: 403 });
        }

        const validSquad = user.activeInternalSquads && user.activeInternalSquads.some(s => ALLOWED_SQUADS.includes(s.uuid));
        if (!validSquad) {
            return new Response(JSON.stringify({ error: "Invalid product/squad for this service" }), { status: 403 });
        }

        // CRITICAL: Check for ANY global pending requests first - this ensures only one request at a time
        const list = await env.REQUESTS.list();
        for (const key of list.keys) {
            const r = await env.REQUESTS.get(key.name, { type: "json" });
            if (r && (r.status === "PENDING" || r.status === "CHANGING")) {
                // 构造Telegram消息链接
                let messageLink = null;
                if (r.message_id && env.TG_CHAT_ID) {
                    // 将 chat_id 转换为 Telegram 链接格式（去掉 -100 前缀）
                    const chatIdStr = String(env.TG_CHAT_ID);
                    const linkChatId = chatIdStr.startsWith('-100') ? chatIdStr.slice(4) : chatIdStr.replace('-', '');
                    messageLink = `https://t.me/c/${linkChatId}/${r.message_id}`;
                }
                return new Response(JSON.stringify({
                    error: "系统当前有待处理的请求，请稍后再试。",
                    message_link: messageLink
                }), { status: 429 });
            }
        }

        // Check user-specific cooldown (only for completed requests)
        const kvKey = `req_${user.username || shortUuid}`;
        const existing = await env.REQUESTS.get(kvKey, { type: "json" });

        if (existing && existing.status === 'COMPLETED') {
            const now = Date.now();
            // 使用 completedAt（完成时间）计算冷却，如果不存在则回退到 timestamp
            const completionTime = existing.completedAt || existing.timestamp;
            const diffHours = (now - completionTime) / (1000 * 60 * 60);
            if (diffHours < COOLDOWN_HOURS) {
                return new Response(JSON.stringify({ error: "Cooldown active. Please wait 6 hours between IP swaps." }), { status: 403 });
            }
        }

        const tgMessage = `<b>🔄 IP 更换请求</b>

<b>👤 发起用户:</b> <code>${user.username}</code>
<b>🕒 提交时间:</b> ${new Date().toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })}
<b>📝 更换原因:</b>
<blockquote>${reason}</blockquote>

<b>📊 当前状态:</b> ⏳ <b>处理中</b> (0/5)`;

        const buttons = {
            inline_keyboard: [[
                { text: "😀 同意 (0)", callback_data: `agree:${user.username}` },
                { text: "😡 拒绝 (0)", callback_data: `decline:${user.username}` }
            ]]
        };

        const tgResp = await fetch(`https://api.telegram.org/bot${env.TG_BOT_TOKEN}/sendMessage`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                chat_id: env.TG_CHAT_ID,
                text: tgMessage,
                parse_mode: "HTML",
                reply_markup: buttons
            })
        });

        let messageId = null;
        if (tgResp.ok) {
            const tgData = await tgResp.json();
            messageId = tgData.result.message_id;
        }

        const newRequest = {
            timestamp: Date.now(),
            username: user.username,
            reason: reason,
            shortUuid: shortUuid,
            count: 0,
            declines: 0,
            status: "PENDING",
            voted_users: [],
            message_id: messageId
        };
        await env.REQUESTS.put(kvKey, JSON.stringify(newRequest));

        return new Response(JSON.stringify({ success: true }));

    } catch (e) {
        return new Response(JSON.stringify({ error: e.message }), { status: 500 });
    }
}

async function handleWebhook(request, env) {
    try {
        const update = await request.json();
        if (!update.callback_query) return new Response("OK");

        const cb = update.callback_query;
        const data = cb.data;
        const [action, username] = data.split(":");
        const userId = cb.from.id;
        const kvKey = `req_${username}`;

        const reqData = await env.REQUESTS.get(kvKey, { type: "json" });
        if (!reqData) {
            await answerCallback(env, cb.id, "请求不存在或已过期");
            return new Response("OK");
        }

        if (reqData.voted_users && reqData.voted_users.includes(userId)) {
            await answerCallback(env, cb.id, "你已经投过票了捏");
            return new Response("OK");
        }

        if (!reqData.voted_users) reqData.voted_users = [];
        reqData.voted_users.push(userId);

        let shouldDelete = false;
        let finalStatus = "PENDING";

        if (action === "agree") {
            reqData.count = (reqData.count || 0) + 1;
            await answerCallback(env, cb.id, "Voted Agree");

            // 当 count 达到 5 时，状态改为 CHANGING
            if (reqData.count >= 5) {
                reqData.status = "CHANGING";
                finalStatus = "CHANGING";
            }
        } else if (action === "decline") {
            reqData.declines = (reqData.declines || 0) + 1;
            await answerCallback(env, cb.id, "Voted Decline");
            if (reqData.declines >= 2) {
                shouldDelete = true;
                finalStatus = "REJECTED";
            }
        }

        if (shouldDelete) {
            await env.REQUESTS.delete(kvKey);
        } else {
            await env.REQUESTS.put(kvKey, JSON.stringify(reqData));
        }

        const timeStr = new Date(reqData.timestamp).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' });

        let messageBody = `<b>🔄 IP 更换请求</b>

<b>👤 发起用户:</b> <code>${reqData.username}</code>
<b>🕒 提交时间:</b> ${timeStr}
<b>📝 更换原因:</b>
<blockquote>${reqData.reason}</blockquote>`;

        if (finalStatus === "REJECTED") {
            messageBody += `\n\n<b>📊 当前状态:</b> ❌ <b>不予通过 (Rejected)</b>`;
            await editMessage(env, cb.message.chat.id, cb.message.message_id, messageBody, null);
        } else if (finalStatus === "CHANGING") {
            messageBody += `\n\n<b>📊 当前状态:</b> 🔄 <b>正在更换IP中 (Changing)</b> (5/5)`;
            await editMessage(env, cb.message.chat.id, cb.message.message_id, messageBody, null);
        } else {
            messageBody += `\n\n<b>📊 当前状态:</b> ⏳ <b>处理中</b> (${reqData.count}/5)`;
            const buttons = {
                inline_keyboard: [[
                    { text: `😀 同意 (${reqData.count || 0})`, callback_data: `agree:${username}` },
                    { text: `😡 拒绝 (${reqData.declines || 0})`, callback_data: `decline:${username}` }
                ]]
            };
            await editMessage(env, cb.message.chat.id, cb.message.message_id, messageBody, buttons);
        }

        return new Response("OK");
    } catch (e) {
        console.error(e);
        return new Response("Error", { status: 500 });
    }
}

async function handleLookup(request, env) {
    const list = await env.REQUESTS.list();

    let changingRequest = null;
    let pendingRequest = null;

    // Loop through all keys to find CHANGING or PENDING requests
    for (const key of list.keys) {
        const r = await env.REQUESTS.get(key.name, { type: "json" });
        // 查找 CHANGING 状态的请求（最高优先级）
        if (r && r.status === "CHANGING") {
            changingRequest = r;
            break; // CHANGING 优先级最高，找到就退出
        }
        // 查找 PENDING 且 count >= 1 的请求（备选）
        if (!pendingRequest && r && r.status === "PENDING" && r.count >= 1) {
            pendingRequest = r;
        }
    }

    // 优先返回 CHANGING 请求，其次返回 PENDING 请求
    const readyRequest = changingRequest || pendingRequest;

    if (readyRequest) {
        return new Response(JSON.stringify({
            count: readyRequest.count,
            status: readyRequest.status
        }));
    } else {
        return new Response(JSON.stringify({
            count: 0,
            status: "WAITING"
        }));
    }
}

async function handleSwap(request, env) {
    const auth = request.headers.get("Authorization");

    if (auth !== SWAP_TOKEN && auth !== `Bearer ${SWAP_TOKEN}`) {
        return new Response("Unauthorized", { status: 401 });
    }

    const list = await env.REQUESTS.list();
    let targetKey = null;
    let targetReq = null;

    // 查找 CHANGING 状态的请求
    for (const key of list.keys) {
        const r = await env.REQUESTS.get(key.name, { type: "json" });
        if (r && r.status === "CHANGING") {
            targetKey = key.name;
            targetReq = r;
            break;
        }
    }

    if (targetKey && targetReq) {
        targetReq.status = "COMPLETED";
        targetReq.count = 0;
        // 更新 timestamp 为完成时间，用于正确计算冷却时间
        targetReq.completedAt = Date.now();

        // Update Telegram Message
        if (targetReq.message_id) {
            const timeStr = new Date(targetReq.timestamp).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' });
            const text = `<b>🔄 IP 更换请求</b>

<b>👤 发起用户:</b> <code>${targetReq.username}</code>
<b>🕒 提交时间:</b> ${timeStr}
<b>📝 更换原因:</b>
<blockquote>${targetReq.reason}</blockquote>

<b>📊 当前状态:</b> ✅ <b>更换IP成功 (IP Swapped)</b>`;
            await editMessage(env, env.TG_CHAT_ID, targetReq.message_id, text, null);
        }

        // We keep the request in KV to serve as cooldown record.
        await env.REQUESTS.put(targetKey, JSON.stringify(targetReq));
        return new Response("OK");
    }

    return new Response("No pending tasks", { status: 404 });
}

// API接口：直接添加换IP请求，需要token验证，不受cooldown限制
async function handleApiAddRequest(request, env) {
    try {
        // Token验证
        const auth = request.headers.get("Authorization");
        if (auth !== SWAP_TOKEN && auth !== `Bearer ${SWAP_TOKEN}`) {
            return new Response(JSON.stringify({ success: false, error: "Unauthorized" }), {
                status: 401,
                headers: { "Content-Type": "application/json" }
            });
        }

        // 获取请求体中的reason（可选）
        let reason = "API自动触发";
        try {
            const body = await request.json();
            if (body.reason) {
                reason = body.reason;
            }
        } catch (e) {
            // 如果没有请求体或解析失败，使用默认reason
        }

        // 检查是否有任何待处理的请求（PENDING 或 CHANGING）
        const list = await env.REQUESTS.list();
        for (const key of list.keys) {
            const r = await env.REQUESTS.get(key.name, { type: "json" });
            if (r && (r.status === "PENDING" || r.status === "CHANGING")) {
                return new Response(JSON.stringify({
                    success: false,
                    error: "已有待处理的换IP请求",
                    existing_status: r.status,
                    existing_username: r.username
                }), {
                    status: 409,
                    headers: { "Content-Type": "application/json" }
                });
            }
        }

        // 创建新的换IP请求，设置为PENDING状态，需要投票
        const username = "API_AUTO";
        const kvKey = `req_${username}`;  // 使用与 webhook 回调匹配的 key 格式

        // 发送Telegram通知（带投票按钮）
        const tgMessage = `<b>🔄 IP 更换请求 (API)</b>

<b>👤 发起用户:</b> <code>${username}</code>
<b>🕒 提交时间:</b> ${new Date().toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })}
<b>📝 更换原因:</b>
<blockquote>${reason}</blockquote>

<b>📊 当前状态:</b> ⏳ <b>处理中</b> (0/5)`;

        const buttons = {
            inline_keyboard: [[
                { text: "😀 同意 (0)", callback_data: `agree:${username}` },
                { text: "😡 拒绝 (0)", callback_data: `decline:${username}` }
            ]]
        };

        const tgResp = await fetch(`https://api.telegram.org/bot${env.TG_BOT_TOKEN}/sendMessage`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                chat_id: env.TG_CHAT_ID,
                text: tgMessage,
                parse_mode: "HTML",
                reply_markup: buttons
            })
        });

        let messageId = null;
        if (tgResp.ok) {
            const tgData = await tgResp.json();
            messageId = tgData.result.message_id;
        }

        const newRequest = {
            timestamp: Date.now(),
            username: username,
            reason: reason,
            shortUuid: "api",
            count: 0,  // 从0开始，需要投票
            declines: 0,
            status: "PENDING",  // 设置为PENDING状态，需要投票通过
            voted_users: [],
            message_id: messageId
        };
        await env.REQUESTS.put(kvKey, JSON.stringify(newRequest));

        return new Response(JSON.stringify({
            success: true,
            message: "换IP请求已添加，等待投票",
            request_id: kvKey
        }), {
            status: 200,
            headers: { "Content-Type": "application/json" }
        });

    } catch (e) {
        return new Response(JSON.stringify({ success: false, error: e.message }), {
            status: 500,
            headers: { "Content-Type": "application/json" }
        });
    }
}

async function answerCallback(env, id, text) {
    await fetch(`https://api.telegram.org/bot${env.TG_BOT_TOKEN}/answerCallbackQuery`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ callback_query_id: id, text: text })
    });
}

async function editMessage(env, chatId, msgId, text, replyMarkup) {
    try {
        const body = {
            chat_id: chatId,
            message_id: msgId,
            text: text,
            parse_mode: "HTML"
        };
        if (replyMarkup) body.reply_markup = replyMarkup;

        const response = await fetch(`https://api.telegram.org/bot${env.TG_BOT_TOKEN}/editMessageText`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body)
        });

        if (!response.ok) {
            console.error(`editMessage failed: ${response.status} ${response.statusText}`);
        }
    } catch (e) {
        console.error(`editMessage error: ${e.message}`);
    }
}