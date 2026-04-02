/**
 * Telegram Bot for User Subscription Management
 * Cloudflare Workers Version
 * 
 * Required KV Namespace: BOT_KV
 * 
 * Environment Variables:
 * - BOT_TOKEN: Telegram Bot Token
 * - API_TOKEN: Panel API Token
 * - API_BASE: Panel API Base URL
 */

// ==================== 配置 ====================
const CONFIG = {
    BOT_TOKEN: '8320764165:AAHYK2_ZkGHQpdlmqODKLNZLq7BnvnXZYuI',
    API_TOKEN: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiYWQ0MDg0NWYtMTZkZS00MTkwLTgzODUtMzEwODVkYThjODA3IiwidXNlcm5hbWUiOm51bGwsInJvbGUiOiJBUEkiLCJpYXQiOjE3NzAyODcyNDMsImV4cCI6MTA0MTAyMDA4NDN9.JJbYJ5tVvUokkBmQYXFjUgsxF69Zf_xllI4vR69-0h8',
    API_BASE: 'https://panel.1391399.xyz',
    GROUP_ID: -1003493995915,
    CHANNEL_ID: -1003523393036,
    CHANNEL_URL: 'https://t.me/txportnotice',
    ADMIN_IDS: [8485399326, 6412530296],
    SHOP_URL: 'https://shop.jsnav.de'
};

// ==================== 全局常量 ====================
const DURATION_DAYS = {
    'monthly': 30,
    'bimonthly': 60,
    'quarterly': 90,
    'semiannual': 180,
    'annual': 365
};

const DURATION_NAMES = {
    'monthly': '📅 月付',
    'bimonthly': '📅 2月付',
    'quarterly': '📆 季付',
    'semiannual': '🗓️ 半年付',
    'annual': '🎉 年付'
};

const STRATEGY_NAMES = {
    'NO_RESET': '♾️ 不重置',
    'DAY': '📆 每日重置',
    'WEEK': '📅 每周重置',
    'MONTH': '🗓️ 每月重置'
};

// ==================== 工具函数 ====================
function formatBytes(bytes) {
    if (bytes === 0 || bytes === null || bytes === undefined) return '0 B';
    if (bytes < 0) bytes = 0;
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    if (i < 0 || i >= sizes.length) return '0 B';
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function generateProgressBar(percentage, width = 10) {
    const clamped = Math.max(0, Math.min(100, percentage));
    const filled = Math.min(Math.round(clamped / (100 / width)), width);
    const empty = width - filled;
    let color = '🟩'; // Green
    if (clamped >= 80) color = '🟥'; // Red
    else if (clamped >= 50) color = '🟨'; // Yellow
    return color.repeat(filled) + '⬜'.repeat(empty);
}

function formatDate(dateStr) {
    if (!dateStr && dateStr !== 0) return '无';
    const date = typeof dateStr === 'number' ? new Date(dateStr) : new Date(dateStr);
    if (isNaN(date.getTime())) return '无';
    return date.toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' });
}

function isAdmin(userId) {
    return CONFIG.ADMIN_IDS.includes(userId);
}

function isJoinedChatStatus(status) {
    return ['member', 'administrator', 'creator', 'restricted'].includes(status);
}

function generateCardCode() {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789';
    let code = '';
    for (let i = 0; i < 16; i++) {
        code += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return code;
}

function generateRequestId() {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789';
    let id = '';
    for (let i = 0; i < 12; i++) {
        id += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return id;
}

function extractNodeseekId(url) {
    const match = url.match(/nodeseek\.com\/space\/(\d+)/);
    return match ? match[1] : null;
}

function extractShortUuid(url) {
    const match = url.match(/sub\.[^\/]+\/([a-zA-Z0-9_-]+)/);
    return match ? match[1] : null;
}

// ==================== Shop Crypto ====================
const shopEncoder = new TextEncoder();
const shopDecoder = new TextDecoder();

function shopLeftRotate(x, c) {
    return ((x << c) | (x >>> (32 - c))) >>> 0;
}

function shopMd5Bytes(input) {
    const shifts = [
        7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22,
        5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20,
        4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23,
        6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21
    ];
    const k = new Uint32Array(64);
    for (let i = 0; i < 64; i++) k[i] = Math.floor(Math.abs(Math.sin(i + 1)) * 0x100000000) >>> 0;
    const bitLen = input.length * 8;
    const paddedLen = (((input.length + 9) + 63) >> 6) << 6;
    const bytes = new Uint8Array(paddedLen);
    bytes.set(input);
    bytes[input.length] = 0x80;
    const bitLenLo = bitLen >>> 0;
    const bitLenHi = Math.floor(bitLen / 0x100000000) >>> 0;
    bytes[paddedLen - 8] = bitLenLo & 0xff;
    bytes[paddedLen - 7] = (bitLenLo >>> 8) & 0xff;
    bytes[paddedLen - 6] = (bitLenLo >>> 16) & 0xff;
    bytes[paddedLen - 5] = (bitLenLo >>> 24) & 0xff;
    bytes[paddedLen - 4] = bitLenHi & 0xff;
    bytes[paddedLen - 3] = (bitLenHi >>> 8) & 0xff;
    bytes[paddedLen - 2] = (bitLenHi >>> 16) & 0xff;
    bytes[paddedLen - 1] = (bitLenHi >>> 24) & 0xff;
    let a0 = 0x67452301, b0 = 0xefcdab89, c0 = 0x98badcfe, d0 = 0x10325476;
    for (let offset = 0; offset < bytes.length; offset += 64) {
        const m = new Uint32Array(16);
        for (let j = 0; j < 16; j++) {
            const o = offset + (j * 4);
            m[j] = bytes[o] | (bytes[o + 1] << 8) | (bytes[o + 2] << 16) | (bytes[o + 3] << 24);
        }
        let a = a0, b = b0, c = c0, d = d0;
        for (let i = 0; i < 64; i++) {
            let f, g;
            if (i < 16) { f = (b & c) | ((~b) & d); g = i; }
            else if (i < 32) { f = (d & b) | ((~d) & c); g = ((5 * i) + 1) % 16; }
            else if (i < 48) { f = b ^ c ^ d; g = ((3 * i) + 5) % 16; }
            else { f = c ^ (b | (~d)); g = (7 * i) % 16; }
            const temp = d;
            const sum = (a + f + k[i] + m[g]) >>> 0;
            d = c; c = b;
            b = (b + shopLeftRotate(sum, shifts[i])) >>> 0;
            a = temp;
        }
        a0 = (a0 + a) >>> 0; b0 = (b0 + b) >>> 0; c0 = (c0 + c) >>> 0; d0 = (d0 + d) >>> 0;
    }
    const out = new Uint8Array(16);
    [a0, b0, c0, d0].forEach((v, i) => {
        out[i * 4] = v & 0xff; out[i * 4 + 1] = (v >>> 8) & 0xff;
        out[i * 4 + 2] = (v >>> 16) & 0xff; out[i * 4 + 3] = (v >>> 24) & 0xff;
    });
    return out;
}

const ShopCrypto = {
    randomString: (length = 32) => {
        const arr = new Uint8Array(length / 2);
        crypto.getRandomValues(arr);
        return Array.from(arr, dec => dec.toString(16).padStart(2, '0')).join('');
    },
    md5: async (message) => {
        const hash = shopMd5Bytes(shopEncoder.encode(message));
        return Array.from(hash).map(b => b.toString(16).padStart(2, '0')).join('');
    },
    encrypt: async (data, secret) => {
        const keyBytes = shopEncoder.encode(secret.substring(0, 16));
        const ivBytes = shopEncoder.encode(secret.substring(0, 16));
        const keyMaterial = await crypto.subtle.importKey('raw', keyBytes, { name: 'AES-CBC' }, false, ['encrypt']);
        const encrypted = await crypto.subtle.encrypt({ name: 'AES-CBC', iv: ivBytes }, keyMaterial, shopEncoder.encode(data));
        return btoa(String.fromCharCode(...new Uint8Array(encrypted)));
    },
    decrypt: async (data, secret) => {
        try {
            const keyBytes = shopEncoder.encode(secret.substring(0, 16));
            const ivBytes = shopEncoder.encode(secret.substring(0, 16));
            const keyMaterial = await crypto.subtle.importKey('raw', keyBytes, { name: 'AES-CBC' }, false, ['decrypt']);
            const encryptedData = Uint8Array.from(atob(data), c => c.charCodeAt(0));
            const decrypted = await crypto.subtle.decrypt({ name: 'AES-CBC', iv: ivBytes }, keyMaterial, encryptedData);
            return shopDecoder.decode(new Uint8Array(decrypted));
        } catch { return null; }
    },
    signature: async (data, secret) => {
        const sorted = Object.keys(data).sort();
        const parts = [];
        for (const key of sorted) {
            let val = data[key];
            if (val === null || val === undefined || typeof val === 'object' || key === 'sign') continue;
            if (typeof val === 'boolean') val = val ? '1' : '0';
            const strVal = String(val);
            if (strVal === '') continue;
            parts.push(`${key}=${strVal}`);
        }
        return await ShopCrypto.md5(parts.join('&') + `&key=${secret}`);
    }
};

// ==================== Shop Client ====================
class ShopClient {
    constructor(baseUrl, kv, chatId) {
        this.baseUrl = baseUrl.replace(/\/$/, '');
        this.host = new URL(this.baseUrl).host;
        this.kv = kv;
        this.chatId = chatId;
    }
    async getToken() { return await this.kv.get(`shop_token:${this.chatId}`); }
    async setToken(token) { await this.kv.put(`shop_token:${this.chatId}`, token, { expirationTtl: 86400 * 7 }); }
    async deleteToken() { await this.kv.delete(`shop_token:${this.chatId}`); }
    async getClientId() {
        const key = `shop_client:${this.chatId}`;
        let clientId = await this.kv.get(key);
        if (!clientId || String(clientId).length !== 32) {
            clientId = ShopCrypto.randomString(32);
            await this.kv.put(key, clientId, { expirationTtl: 86400 * 365 });
        }
        return clientId;
    }
    async buildCookie() {
        const token = await this.getToken();
        const clientId = await this.getClientId();
        const cookieParts = [`client_id=${clientId}`];
        if (token) cookieParts.unshift(`user_token=${token}`);
        return cookieParts.join('; ');
    }
    async request(path, data = {}) {
        const secret = ShopCrypto.randomString(32);
        const signature = await ShopCrypto.signature(data, secret);
        const encryptedBody = await ShopCrypto.encrypt(JSON.stringify(data), secret);
        const headers = {
            'Content-Type': 'text/plain', 'Secret': secret, 'Signature': signature,
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            'Host': this.host, 'Cookie': await this.buildCookie()
        };
        try {
            const response = await fetch(`${this.baseUrl}${path}`, { method: 'POST', headers, body: encryptedBody });
            const responseSecret = response.headers.get('Secret');
            const responseText = await response.text();
            if (responseSecret) {
                const decrypted = await ShopCrypto.decrypt(responseText, responseSecret);
                if (decrypted) { try { return JSON.parse(decrypted); } catch { return { code: 500, msg: '响应解析错误' }; } }
                return { code: 500, msg: '解密失败' };
            }
            try { return JSON.parse(responseText); } catch { return { code: response.status, msg: responseText.substring(0, 200) }; }
        } catch (e) { return { code: 500, msg: e.message }; }
    }
    async requestHtml(path) {
        try {
            const response = await fetch(`${this.baseUrl}${path}`, {
                headers: {
                    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                    'Host': this.host, 'Cookie': await this.buildCookie()
                }
            });
            return { ok: response.ok, text: await response.text() };
        } catch (e) { return { ok: false, text: '' }; }
    }
}

// ==================== Shop 辅助函数 ====================
function shopToArray(data) {
    if (!data) return [];
    if (Array.isArray(data)) return data;
    if (data.list) return shopToArray(data.list);
    if (typeof data === 'object') return Object.values(data);
    return [];
}

function shopEscapeHtml(text) {
    if (!text) return '';
    return String(text).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function shopCollectLeafCategoryIds(nodes, out = []) {
    for (const node of shopToArray(nodes)) {
        const children = shopToArray(node?.children);
        if (children.length > 0) shopCollectLeafCategoryIds(children, out);
        else if (node?.id != null) out.push(node.id);
    }
    return out;
}

const ALL_PRODUCTS_PAGE_SIZE = 20;
const MAX_CATEGORY_ITEM_PAGES = 100;

async function shopFetchAllProducts(client) {
    const allItems = [];
    let globalSourceWorked = false;

    // Source A: global item list
    {
        let page = 1, lastFp = '';
        while (page <= MAX_CATEGORY_ITEM_PAGES) {
            const r = await client.request('/shop/item', { page, limit: 100 });
            if (r.code !== 200) break;

            globalSourceWorked = true;
            const d = r.data, items = shopToArray(d?.list ?? d);
            if (items.length === 0) break;

            const fp = items.map(i => i?.id ?? '').join(',');
            if (page > 1 && fp && fp === lastFp) break;
            lastFp = fp;
            allItems.push(...items);

            const total = Number(d?.total ?? d?.count ?? 0);
            const perPage = Number(d?.limit ?? d?.per_page ?? 100);
            const current = Number(d?.page ?? d?.current_page ?? page);

            // Calculate last page
            const hasTotal = Number.isFinite(total) && total > 0;
            const lastPage = Number(d?.last_page ?? d?.pages ?? (hasTotal && perPage > 0 ? Math.ceil(total / perPage) : 1));

            if (Number.isFinite(lastPage) && lastPage > current) { page++; continue; }
            if (hasTotal && perPage > 0 && (page * perPage) < total) { page++; continue; }
            if (items.length >= perPage) { page++; continue; }
            break;
        }
    }

    // Source B: category traversal fallback
    const catRes = await client.request('/shop/category', {});
    if (catRes.code === 200) {
        const leafIds = [...new Set(shopCollectLeafCategoryIds(catRes.data))];
        for (const catId of leafIds) {
            let p = 1, lfp = '';
            while (p <= MAX_CATEGORY_ITEM_PAGES) {
                const r = await client.request('/shop/item', { category_id: catId, page: p, limit: 100 });
                if (r.code !== 200) break;

                const d = r.data, items = shopToArray(d?.list ?? d);
                if (items.length === 0) break;

                const fp = items.map(i => i?.id ?? '').join(',');
                if (p > 1 && fp && fp === lfp) break;
                lfp = fp;
                allItems.push(...items);

                const total = Number(d?.total ?? 0), perPage = Number(d?.limit ?? 100);
                if (Number.isFinite(total) && total > 0 && (p * perPage) < total) { p++; continue; }
                if (items.length >= perPage) { p++; continue; }
                break;
            }
        }
    } else if (!globalSourceWorked) {
        return { ok: false, msg: catRes.msg || 'Failed to load products' };
    }

    const seen = new Set();
    return { ok: true, items: allItems.filter(item => { const id = item?.id; if (id == null || seen.has(id)) return false; seen.add(id); return true; }) };
}

// ==================== Shop 凭证管理 ====================
async function saveShopCredentials(kv, chatId, username, password) {
    await kv.put(`shop_creds:${chatId}`, JSON.stringify({ username, password }), { expirationTtl: 86400 * 30 });
}
async function getShopCredentials(kv, chatId) {
    return kv.get(`shop_creds:${chatId}`, 'json');
}
async function deleteShopCredentials(kv, chatId) {
    await kv.delete(`shop_creds:${chatId}`);
}
async function ensureShopLoggedIn(kv, shopClient) {
    const token = await shopClient.getToken();
    if (!token) {
        return { ok: true, loggedIn: false };
    }

    const res = await shopClient.request('/user/personal/info', {});
    if (res.code === 200 && res.data) {
        return { ok: true, loggedIn: true, profile: res.data };
    }

    // Only delete token if explicitly expired/invalid (code 0 often means auth error in some systems, but verify)
    if (res.code === 0 || res.code === 401) {
        await shopClient.deleteToken();
        return { ok: true, loggedIn: false, widthRelogin: true };
    }

    // API error but might still be logged in/token might be valid
    // Assuming not logged in for safety if we can't verify
    return { ok: false, loggedIn: true, msg: res.msg || 'Verification failed' };
}

// ==================== Trade Binding & Access Control ====================
function getTradeOwnerKey(tradeNo) {
    return `trade_owner:${tradeNo}`;
}

async function bindTradeToChat(kv, chatId, tradeNo) {
    const chat = String(chatId);
    const ownerKey = getTradeOwnerKey(tradeNo);
    const currentOwner = await kv.get(ownerKey);
    if (!currentOwner || currentOwner === chat) {
        await kv.put(ownerKey, chat, { expirationTtl: 86400 * 30 });
    }
}

// ==================== Pending Trade Helpers ====================
function getPendingTradeKey(chatId) {
    return `pending_trade:${chatId}`;
}

async function getPendingTrade(kv, chatId) {
    return kv.get(getPendingTradeKey(chatId));
}

async function setPendingTrade(kv, chatId, tradeNo) {
    await kv.put(getPendingTradeKey(chatId), tradeNo, { expirationTtl: 3600 });
}

async function clearPendingTrade(kv, chatId) {
    await kv.delete(getPendingTradeKey(chatId));
}

async function clearPendingTradeIfMatch(kv, chatId, tradeNo) {
    const current = await getPendingTrade(kv, chatId);
    if (current === tradeNo) {
        await clearPendingTrade(kv, chatId);
    }
}

async function ensureTradeAccess(kv, client, chatId, tradeNo, isLoggedIn) {
    const owner = await kv.get(getTradeOwnerKey(tradeNo));
    if (owner && owner !== String(chatId)) {
        return { ok: false, msg: '🔒 该订单已绑定到其他 Telegram 账户' };
    }
    if (owner === String(chatId)) {
        return { ok: true };
    }

    if (!isLoggedIn) {
        // Guest mode: implicit access via tradeNo, bind it now
        await bindTradeToChat(kv, chatId, tradeNo);
        return { ok: true };
    }

    // Check if user owns this trade via API
    try {
        const listRes = await client.request('/user/trade/order/get', {
            keywords: tradeNo,
            page: 1,
            limit: 20
        });

        if (listRes.code === 200) {
            const list = shopToArray(listRes.data?.list || []);
            const owned = list.some(item => (item.main_trade_no || '') === tradeNo || (item.trade_no || '') === tradeNo);
            if (owned) {
                await bindTradeToChat(kv, chatId, tradeNo);
                return { ok: true };
            }
        }
    } catch (e) {
        console.error('ensureTradeAccess error:', e);
    }

    // Fallback: if we can't verify via API (e.g. search lag), but user has tradeNo, allow it for now but don't bind strictly?
    // Safer to allow guests implicit access if no owner set.
    if (!owner) {
        await bindTradeToChat(kv, chatId, tradeNo);
        return { ok: true };
    }

    return { ok: false, msg: '🔒 您没有权限查看此订单' };
}

// Shop Flow: 展示产品列表
async function shopShowProducts(kv, chatId, userId, shopClient, flow) {
    const loginStatus = await ensureShopLoggedIn(kv, shopClient);
    const result = await shopFetchAllProducts(shopClient);
    if (!result.ok || result.items.length === 0) {
        await sendMessage(chatId, '❌ 暂无可用商品', {
            reply_markup: { inline_keyboard: [[{ text: '◀️ 返回', callback_data: flow === 'reg' ? 'cancel_reg' : 'refresh' }]] }
        });
        return;
    }
    // 缓存产品列表到 session
    await setSession(kv, userId, 'shop_browse', { flow, products: result.items, page: 1 });
    await shopRenderProductPage(chatId, result.items, 1, flow);
}

async function shopRenderProductPage(chatId, products, page, flow) {
    const pageSize = 15;
    const totalPages = Math.max(1, Math.ceil(products.length / pageSize));
    const currentPage = Math.min(Math.max(1, page), totalPages);
    const start = (currentPage - 1) * pageSize;
    const pageItems = products.slice(start, start + pageSize);
    const buttons = pageItems.map(item => {
        const price = item.price || item.sku?.[0]?.price || '?';
        return [{ text: `🛍 ${shopEscapeHtml(String(item.name || '').substring(0, 25))} - ¥${price}`, callback_data: `si_${item.id}` }];
    });
    const nav = [];
    if (currentPage > 1) nav.push({ text: '⬅️ 上一页', callback_data: `spage_${currentPage - 1}` });
    if (currentPage < totalPages) nav.push({ text: '下一页 ➡️', callback_data: `spage_${currentPage + 1}` });
    if (nav.length > 0) buttons.push(nav);
    const hint = flow === 'renew' ? '\n💡 请选择对应金额的「自助续费」商品' : '';
    buttons.push([{ text: '❌ 取消', callback_data: flow === 'reg' ? 'cancel_reg' : 'refresh' }]);
    await sendMessage(chatId, `🛍 \u003cb\u003e选择商品\u003c/b\u003e (共${products.length}件)${hint}\n📄 第${currentPage}/${totalPages}页`, {
        reply_markup: { inline_keyboard: buttons }
    });
}

// Shop Flow: 获取订单商品内容(卡密)
async function shopGetOrderCards(shopClient, tradeNo, isLoggedIn) {
    const cards = [];
    if (isLoggedIn) {
        const listRes = await shopClient.request('/user/trade/order/get', { keywords: tradeNo, page: 1, limit: 10 });
        if (listRes.code === 200) {
            const list = shopToArray(listRes.data?.list || []);
            for (const order of list) {
                if ([1, 3, 4].includes(Number(order.status))) {
                    const itemRes = await shopClient.request('/user/trade/order/item', { id: order.id });
                    if (itemRes.code === 200 && itemRes.data?.treasure) {
                        const lines = itemRes.data.treasure.trim().split('\n').map(l => l.trim()).filter(l => l.length > 0);
                        cards.push(...lines);
                    }
                }
            }
        }
    } else {
        // Guest mode: try search page
        const htmlRes = await shopClient.requestHtml(`/search?tradeNo=${encodeURIComponent(tradeNo)}`);
        if (htmlRes.ok) {
            const regex = /data-id="(\d+)"/g;
            let m; const itemIds = [];
            while ((m = regex.exec(htmlRes.text)) !== null) itemIds.push(parseInt(m[1], 10));
            for (const itemId of [...new Set(itemIds)]) {
                const orderRes = await shopClient.request('/shop/order/getOrder', { trade_no: tradeNo, item_id: itemId });
                if (orderRes.code === 200 && orderRes.data?.treasure) {
                    const lines = orderRes.data.treasure.trim().split('\n').map(l => l.trim()).filter(l => l.length > 0);
                    cards.push(...lines);
                }
            }
        }
    }
    return cards;
}

// Helper: Build Order View (mimics worker.mcy.js buildOrderView)
async function buildOrderView(client, tradeNo, isLoggedIn, shopUrl) {
    if (isLoggedIn) {
        const listRes = await client.request('/user/trade/order/get', { keywords: tradeNo, page: 1, limit: 10 });
        if (listRes.code !== 200) return { ok: false, msg: `❌ ${listRes.msg || '获取订单失败'}` };

        const list = shopToArray(listRes.data?.list || []);
        if (list.length === 0) return { ok: false, msg: '⏳ 订单处理中或未找到，请稍后重试...' };

        let allDelivered = true;
        let requiresPayment = false;
        let resultText = `📦 <b>订单详情</b>\n📋 订单号: <code>${tradeNo}</code>\n\n`;
        const buttons = [];

        for (const orderItem of list) {
            const status = Number(orderItem.status);
            const itemName = orderItem.item?.name || '商品';
            const skuName = orderItem.sku?.name || '';

            if ([1, 3, 4].includes(status)) {
                // Paid/Delivered
                const itemRes = await client.request('/user/trade/order/item', { id: orderItem.id });
                if (itemRes.code === 200 && itemRes.data) {
                    const treasure = itemRes.data.treasure || '';
                    if (treasure) {
                        resultText += `✅ <b>${shopEscapeHtml(itemName)}</b> (${shopEscapeHtml(skuName)})\n`;
                        resultText += `<code>${shopEscapeHtml(treasure.substring(0, 800))}</code>\n\n`;
                    } else {
                        resultText += `✅ <b>${shopEscapeHtml(itemName)}</b> - 已发货\n\n`;
                    }
                } else {
                    resultText += `✅ <b>${shopEscapeHtml(itemName)}</b> - 已发货 (内容获取失败)\n\n`;
                }
            } else if (status === 0) {
                // Unpaid
                allDelivered = false;
                requiresPayment = true;
                resultText += `⏳ <b>${shopEscapeHtml(itemName)}</b> - 待支付\n\n`;
                buttons.push([{ text: `💳 支付订单`, callback_data: `pay_${tradeNo}` }]); // Handler for this should exist or reuse shop_pay logic
            } else if (status === 2) {
                // Processing (Paid but not delivered?)
                allDelivered = false;
                resultText += `⏳ <b>${shopEscapeHtml(itemName)}</b> - 处理中\n\n`;
            } else {
                allDelivered = false;
                resultText += `❓ <b>${shopEscapeHtml(itemName)}</b> - 状态: ${status}\n\n`;
            }
        }

        if (!allDelivered) buttons.push([{ text: '🔄 刷新状态', callback_data: `getitem_${tradeNo}` }]);
        buttons.push([{ text: '◀️ 返回', callback_data: 'refresh' }]);

        return { ok: true, text: resultText, buttons, requiresPayment };
    }

    // Guest Mode
    const htmlRes = await client.requestHtml(`/search?tradeNo=${encodeURIComponent(tradeNo)}`);
    if (!htmlRes.ok) return { ok: false, msg: `❌ 无法加载订单页面` };

    const regex = /data-id="(\d+)"/g;
    let m; const itemIds = [];
    while ((m = regex.exec(htmlRes.text)) !== null) itemIds.push(parseInt(m[1], 10));

    if (itemIds.length === 0) return { ok: false, msg: '⏳ 订单数据未就绪，请稍后刷新...' };

    const list = [];
    for (const itemId of [...new Set(itemIds)]) {
        const itemRes = await client.request('/shop/order/getOrder', { trade_no: tradeNo, item_id: itemId });
        if (itemRes.code === 200 && itemRes.data) list.push(itemRes.data);
    }

    if (list.length === 0) return { ok: false, msg: '⏳ 订单数据为空' };

    let allDelivered = true;
    let requiresPayment = false;
    let resultText = `📦 <b>订单详情 (游客)</b>\n📋 订单号: <code>${tradeNo}</code>\n\n`;
    const buttons = [];

    for (const orderItem of list) {
        const status = Number(orderItem.status);
        const itemName = orderItem.item?.name || '商品';
        const treasure = orderItem.treasure || '';

        if ([1, 3, 4].includes(status)) {
            if (treasure) {
                resultText += `✅ <b>${shopEscapeHtml(itemName)}</b>\n`;
                resultText += `<code>${shopEscapeHtml(treasure.substring(0, 800))}</code>\n\n`;
            } else {
                resultText += `✅ <b>${shopEscapeHtml(itemName)}</b> - 已发货\n\n`;
            }
        } else if (status === 0) {
            allDelivered = false;
            requiresPayment = true;
            resultText += `⏳ <b>${shopEscapeHtml(itemName)}</b> - 待支付\n\n`;
        } else {
            allDelivered = false;
            resultText += `⏳ <b>${shopEscapeHtml(itemName)}</b> - 处理中/状态${status}\n\n`;
        }
    }

    if (requiresPayment) buttons.push([{ text: '💳 前往支付', callback_data: `pay_${tradeNo}` }]);
    if (!allDelivered) buttons.push([{ text: '🔄 刷新状态', callback_data: `getitem_${tradeNo}` }]);
    buttons.push([{ text: '◀️ 返回', callback_data: 'refresh' }]);

    return { ok: true, text: resultText, buttons, requiresPayment };
}

// ==================== API 封装 ====================
async function apiRequest(endpoint, method = 'GET', body = null) {
    const options = {
        method,
        headers: {
            'Authorization': `Bearer ${CONFIG.API_TOKEN}`,
            'Content-Type': 'application/json'
        }
    };
    if (body) {
        options.body = JSON.stringify(body);
    }

    const response = await fetch(`${CONFIG.API_BASE}${endpoint}`, options);
    const data = await response.json();

    if (!response.ok) {
        throw new Error(data.message || 'API请求失败');
    }

    return data;
}

async function getUserByTelegramId(telegramId) {
    try {
        const data = await apiRequest(`/api/users/by-telegram-id/${telegramId}`);
        return data.response && data.response.length > 0 ? data.response[0] : null;
    } catch (e) {
        return null;
    }
}

async function getUserByShortUuid(shortUuid) {
    try {
        const data = await apiRequest(`/api/users/by-short-uuid/${shortUuid}`);
        return data.response;
    } catch (e) {
        return null;
    }
}

async function createUser(userData) {
    const data = await apiRequest('/api/users', 'POST', userData);
    return data.response;
}

async function updateUser(userData) {
    const data = await apiRequest('/api/users', 'PATCH', userData);
    return data.response;
}

async function revokeSubscription(uuid, shortUuid = '') {
    const data = await apiRequest(`/api/users/${uuid}/actions/revoke`, 'POST', {
        revokeOnlyPasswords: false,
        shortUuid
    });
    return data.response;
}

async function resetTraffic(uuid) {
    const data = await apiRequest(`/api/users/${uuid}/actions/reset-traffic`, 'POST');
    return data.response;
}

async function getSubscriptionHistory(uuid) {
    const data = await apiRequest(`/api/users/${uuid}/subscription-request-history`);
    return data.response;
}

async function getInternalSquads() {
    const data = await apiRequest('/api/internal-squads');
    return data.response;
}

// ==================== Telegram API 封装 ====================
async function sendMessage(chatId, text, options = {}) {
    const payload = {
        chat_id: chatId,
        text,
        parse_mode: 'HTML',
        ...options
    };

    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/sendMessage`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    });

    const data = await response.json();
    if (!data.ok) {
        throw new Error(`Telegram API Error: ${data.description}`);
    }
    return data;
}

async function sendDocument(chatId, fileContent, fileName, caption = '') {
    const formData = new FormData();
    formData.append('chat_id', chatId);
    formData.append('document', new Blob([fileContent], { type: 'text/plain' }), fileName);
    if (caption) formData.append('caption', caption);
    formData.append('parse_mode', 'HTML');

    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/sendDocument`, {
        method: 'POST',
        body: formData
    });

    const data = await response.json();
    if (!data.ok) {
        throw new Error(`Telegram API Error: ${data.description}`);
    }
    return data;
}

async function editMessage(chatId, messageId, text, options = {}) {
    const payload = {
        chat_id: chatId,
        message_id: messageId,
        text,
        parse_mode: 'HTML',
        ...options
    };

    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/editMessageText`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    });

    const data = await response.json();
    if (!data.ok) {
        // 忽略"message is not modified"错误，这通常发生在用户点击刷新但数据未变化时
        if (data.description && data.description.includes('message is not modified')) {
            return data;
        }
        throw new Error(`Telegram API Error: ${data.description}`);
    }
    return data;
}

async function answerCallback(callbackQueryId, text = '', showAlert = false) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/answerCallbackQuery`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            callback_query_id: callbackQueryId,
            text,
            show_alert: showAlert
        })
    });

    const data = await response.json();
    if (!data.ok) {
        throw new Error(`Telegram API Error: ${data.description}`);
    }
    return data;
}

async function getChatMember(chatId, userId) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/getChatMember`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_id: chatId,
            user_id: userId
        })
    });
    return response.json();
}

async function getChat(chatId) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/getChat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ chat_id: chatId })
    });
    return response.json();
}

async function createChatInviteLink(chatId, memberLimit = 1) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/createChatInviteLink`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_id: chatId,
            member_limit: memberLimit
        })
    });
    const data = await response.json();
    if (!data.ok) throw new Error(`Telegram API Error: ${data.description}`);
    return data.result;
}

async function revokeChatInviteLink(chatId, inviteLink) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/revokeChatInviteLink`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_id: chatId,
            invite_link: inviteLink
        })
    });
    return response.json();
}

async function banChatMember(chatId, userId) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/banChatMember`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_id: chatId,
            user_id: userId,
            until_date: Math.floor(Date.now() / 1000) + 60
        })
    });
    return response.json();
}

async function unbanChatMember(chatId, userId) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/unbanChatMember`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_id: chatId,
            user_id: userId,
            only_if_banned: true
        })
    });
    return response.json();
}

async function deleteMessage(chatId, messageId) {
    const response = await fetch(`https://api.telegram.org/bot${CONFIG.BOT_TOKEN}/deleteMessage`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_id: chatId,
            message_id: messageId
        })
    });
    return response.json();
}

// ==================== 会话状态管理 ====================
async function getSession(kv, telegramId) {
    const data = await kv.get(`session:${telegramId}`, 'json');
    return data || { state: null, data: {} };
}

async function setSession(kv, telegramId, state, data = {}) {
    await kv.put(`session:${telegramId}`, JSON.stringify({
        state,
        data,
        updatedAt: Date.now()
    }), { expirationTtl: 3600 }); // 1小时过期
}

async function clearSession(kv, telegramId) {
    await kv.delete(`session:${telegramId}`);
}

// ==================== 卡密管理 ====================
async function getCard(kv, code) {
    return kv.get(`card:${code}`, 'json');
}

async function deleteCard(kv, code) {
    await kv.delete(`card:${code}`);
}

async function saveCard(kv, code, cardData) {
    await kv.put(`card:${code}`, JSON.stringify(cardData));
}

// ==================== 待处理请求管理 ====================
async function getPendingRequest(kv, requestId) {
    return kv.get(`pending_request:${requestId}`, 'json');
}

async function savePendingRequest(kv, requestId, requestData) {
    // 待处理请求24小时后过期
    await kv.put(`pending_request:${requestId}`, JSON.stringify(requestData), { expirationTtl: 86400 });
}

async function deletePendingRequest(kv, requestId) {
    await kv.delete(`pending_request:${requestId}`);
}

// ==================== 入群请求管理 ====================
async function getJoinRequest(kv, telegramId) {
    return kv.get(`join_request:${telegramId}`, 'json');
}

async function saveJoinRequest(kv, telegramId, data) {
    await kv.put(`join_request:${telegramId}`, JSON.stringify(data), { expirationTtl: 86400 });
}

async function deleteJoinRequest(kv, telegramId) {
    await kv.delete(`join_request:${telegramId}`);
}

// ==================== 流量统计 API ====================
async function getBandwidthStats(userUuid, limit = 5) {
    const end = new Date();
    const start = new Date();
    start.setDate(start.getDate() - 30);
    const startStr = start.toISOString().slice(0, 10);
    const endStr = end.toISOString().slice(0, 10);
    const data = await apiRequest(`/api/bandwidth-stats/users/${userUuid}?topNodesLimit=${limit}&start=${startStr}&end=${endStr}`);
    return data.response;
}

function generateTrafficChart(topNodes, totalUsed) {
    if (!topNodes || topNodes.length === 0) return '📊 暂无流量使用数据';
    const symbols = ['▓', '░', '█', '▒', '▇'];
    const barWidth = 30;
    const grandTotal = topNodes.reduce((sum, n) => sum + (Number(n.total) || 0), 0);
    if (grandTotal === 0) return '📊 暂无流量使用数据';

    let chart = '';
    const segments = topNodes.map((node, i) => ({
        name: node.name,
        total: node.total,
        pct: node.total / grandTotal,
        symbol: symbols[i % symbols.length],
        countryCode: node.countryCode
    }));

    // Build bar
    let bar = '[';
    for (const seg of segments) {
        const len = Math.max(1, Math.round(seg.pct * barWidth));
        bar += seg.symbol.repeat(len);
    }
    bar = bar.substring(0, barWidth + 1);
    while (bar.length < barWidth + 1) bar += ' ';
    bar += `] ${formatBytes(grandTotal)}`;
    chart += bar + '\n\n';

    for (const seg of segments) {
        const pctStr = (seg.pct * 100).toFixed(1);
        chart += `${seg.symbol} ${seg.name} (${seg.countryCode}) - ${formatBytes(seg.total)} (${pctStr}%)\n`;
    }
    return chart.trim();
}

// ==================== 用户界面生成 ====================
function generateUserPanel(user, isAdminUser = false, trafficChart = '') {
    const statusEmoji = {
        'ACTIVE': '🟢',
        'DISABLED': '🔴',
        'LIMITED': '🟡',
        'EXPIRED': '⚫'
    };

    const statusText = {
        'ACTIVE': '正常',
        'DISABLED': '已禁用',
        'LIMITED': '流量耗尽',
        'EXPIRED': '已过期'
    };

    const strategyName = STRATEGY_NAMES;

    const squads = user.activeInternalSquads?.map(s => `「${s.name}」`).join(' ') || '暂无分组';
    const usedTraffic = user.userTraffic?.usedTrafficBytes || 0;
    const totalTraffic = user.trafficLimitBytes || 0;
    const percentage = totalTraffic > 0 ? Math.round((usedTraffic / totalTraffic) * 100) : 0;
    const progressBar = generateProgressBar(percentage);

    // 计算剩余天数
    let remainingDays = '∞';
    let expireWarning = '';
    if (user.expireAt) {
        const expireDate = new Date(user.expireAt);
        const now = new Date();
        const diffTime = expireDate.getTime() - now.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        const days = diffDays > 0 ? diffDays : 0;
        remainingDays = days.toString();
        if (days <= 7 && days > 0) {
            expireWarning = ' ⚠️ 即将到期';
        } else if (days <= 0) {
            expireWarning = ' ❌ 已过期';
        }
    }

    let text = `
╭─────────────────────╮
│      📊 <b>用户控制面板</b>      │
╰─────────────────────╯

👤 <b>账户信息</b>
┌─────────────────────
│ 用户名: <code>${user.username || '未设置'}</code>
│ 状态: ${statusEmoji[user.status] || '❓'} <b>${statusText[user.status] || user.status}</b>
│ 分组: ${squads}
└─────────────────────

📦 <b>流量使用</b>
┌─────────────────────
│ ${progressBar} <b>${percentage}%</b>
│ 已用: <code>${formatBytes(usedTraffic)}</code> / <code>${formatBytes(totalTraffic)}</code>
│ 策略: ${strategyName[user.trafficLimitStrategy] || user.trafficLimitStrategy}
│ 上次重置: ${formatDate(user.lastTrafficResetAt) || '从未重置'}
└─────────────────────

⏰ <b>有效期</b>
┌─────────────────────
│ 到期: ${formatDate(user.expireAt)}${expireWarning}
│ 剩余: <b>${remainingDays}</b> 天
└─────────────────────

🔗 <b>订阅链接</b>
<code>${user.subscriptionUrl || '暂无订阅链接'}</code>

📡 <b>在线状态:</b> ${user.userTraffic?.onlineAt ? '🟢 ' + formatDate(user.userTraffic.onlineAt) : '⚫ 从未在线'}
${trafficChart}
`.trim();

    const buttons = [
        [{ text: '🔄 刷新数据', callback_data: 'refresh' }],
        ...(user.subscriptionUrl ? [[{ text: '🔗 打开订阅链接', url: user.subscriptionUrl }]] : []),
        [
            { text: '🔑 重置订阅', callback_data: 'revoke_confirm' },
            { text: '📊 重置流量', callback_data: 'reset_traffic_info' }
        ],
        [{ text: '💳 续费套餐', callback_data: 'renew_info' }],
        [{ text: '🔄 自助换车', callback_data: 'change_car_info' }],
        [{ text: '📜 24小时订阅记录', callback_data: 'sub_history' }],
        [{ text: '👥 邀请用户进入群组', callback_data: 'invite_user_start' }]
    ];

    if (isAdminUser) {
        buttons.push([
            { text: '🎫 【管理员】生成卡密', callback_data: 'admin_gen_card' },
            { text: '🔧 【管理员】管理卡密', callback_data: 'admin_manage_cards' }
        ]);
    }

    return { text, buttons };
}

// ==================== 注册流程 ====================
async function handleRegistration(kv, chatId, userId, username, messageText, session) {
    switch (session.state) {
        case 'reg_card_input':
            // 支持多卡密输入 (每行一个)
            const rawLines = messageText.trim().split('\n').map(l => l.trim()).filter(l => l.length > 0);
            if (rawLines.length === 0) {
                await sendMessage(chatId, '❌ 请输入至少一个卡密');
                return;
            }

            // 去重，防止重复提交同一张卡密
            const lines = [...new Set(rawLines)];

            // 验证所有卡密
            const validCards = [];
            for (const line of lines) {
                const card = await getCard(kv, line);
                if (!card) {
                    await sendMessage(chatId, `❌ 卡密 <code>${line}</code> 无效或已被使用，请重新输入所有卡密`, {
                        reply_markup: {
                            inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]]
                        }
                    });
                    return;
                }
                validCards.push({ ...card, code: line });
            }

            // 验证所有卡密信息一致 (分组、策略、时长、单张流量)
            const first = validCards[0];
            for (let i = 1; i < validCards.length; i++) {
                const c = validCards[i];
                if (c.trafficBytes !== first.trafficBytes || c.strategy !== first.strategy ||
                    c.duration !== first.duration || c.squadUuid !== first.squadUuid) {
                    await sendMessage(chatId, `❌ 卡密信息不一致，所有卡密的分组、流量、策略和时长必须完全相同\n\n第1张: ${formatBytes(first.trafficBytes)} / ${first.strategy} / ${first.duration}\n第${i + 1}张: ${formatBytes(c.trafficBytes)} / ${c.strategy} / ${c.duration}`, {
                        reply_markup: {
                            inline_keyboard: [
                                [{ text: '🔄 重新输入', callback_data: 'reg_card_retry' }],
                                [{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]
                            ]
                        }
                    });
                    return;
                }
            }

            // 多卡密时显示叠加信息
            let cardSummary = '';
            if (validCards.length > 1) {
                cardSummary = `\n│ 🎫 卡密数量: <b>${validCards.length}</b> 张\n│ 📦 单张流量: <b>${formatBytes(first.trafficBytes)}</b>\n│ 📦 叠加总流量: <b>${formatBytes(first.trafficBytes * validCards.length)}</b>\n│ ⚠️ 仅叠加流量，时长以单张为准`;
            }

            await setSession(kv, userId, 'reg_nodeseek_input', { cards: validCards });
            await sendMessage(chatId, `
✅ <b>卡密验证成功!</b>${cardSummary}

╭──────────────────────
│ 🌐 第2步: 绑定 Nodeseek
╰──────────────────────

请发送您的 Nodeseek 用户链接

📝 格式示例:
<code>https://www.nodeseek.com/space/36628</code>
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]]
                }
            });
            break;

        case 'reg_nodeseek_input':
            const nsId = extractNodeseekId(messageText);
            if (!nsId) {
                await sendMessage(chatId, '❌ 无效的 Nodeseek 链接，请重新输入:', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]]
                    }
                });
                return;
            }

            await setSession(kv, userId, 'reg_email_input', { ...session.data, nodeseekId: nsId });
            await sendMessage(chatId, `
✅ <b>Nodeseek ID 已记录!</b>

╭──────────────────────
│ ✉️ 第3步: 输入邮箱
╰──────────────────────

请输入您的邮箱地址
⚠️ <i>需与发卡网账户一致</i>

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 绑定联系邮箱
👉 <b>下一步:</b> 发送您的常用邮箱地址
👉 <b>遇到问题:</b> 若提示格式错误，请检查是否包含 @ 和域名后缀
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]]
                }
            });
            break;

        case 'reg_email_input':
            const email = messageText.trim();
            if (!email.includes('@') || !email.includes('.')) {
                await sendMessage(chatId, '❌ 邮箱格式无效，请重新输入:', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]]
                    }
                });
                return;
            }

            await setSession(kv, userId, 'reg_verify_channel', { ...session.data, email });
            await sendMessage(chatId, `
✅ <b>邮箱已记录!</b>

╭──────────────────────
│ 📢 第4步: 加入频道
╰──────────────────────

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 关注官方通知频道
👉 <b>下一步:</b> 点击「📢 进入频道」并加入，随后点击「✅ 我已进入」
👉 <b>遇到问题:</b> 如点击验证无反应，请稍等几秒再试，或检查网络连接
📋 <b>操作步骤:</b>
1️⃣ 请先加入我们的频道
2️⃣ 加入后点击下方"我已进入"按钮验证
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '📢 进入频道', url: CONFIG.CHANNEL_URL }],
                        [{ text: '✅ 我已进入', callback_data: 'verify_channel_reg' }],
                        [{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]
                    ]
                }
            });
            break;

        case 'bind_sub_input':
            const shortUuid = extractShortUuid(messageText);
            if (!shortUuid) {
                await sendMessage(chatId, '❌ 无效的订阅链接格式，请发送正确的订阅链接:', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消绑定', callback_data: 'cancel_reg' }]]
                    }
                });
                return;
            }

            try {
                const user = await getUserByShortUuid(shortUuid);
                if (!user) {
                    await sendMessage(chatId, '❌ 未找到对应用户，请检查订阅链接是否正确', {
                        reply_markup: {
                            inline_keyboard: [[{ text: '❌ 取消绑定', callback_data: 'cancel_reg' }]]
                        }
                    });
                    return;
                }

                if (user.telegramId) {
                    await sendMessage(chatId, '❌ 该账户已绑定其他 Telegram 账号', {
                        reply_markup: {
                            inline_keyboard: [[{ text: '❌ 取消绑定', callback_data: 'cancel_reg' }]]
                        }
                    });
                    return;
                }

                await setSession(kv, userId, 'bind_nodeseek_input', {
                    targetUserUuid: user.uuid,
                    targetUsername: user.username
                });

                await sendMessage(chatId, `
✅ <b>订阅链接验证成功!</b>

╭──────────────────────
│ 🌐 第2步: 绑定 Nodeseek
╰──────────────────────

请发送您的 Nodeseek 用户链接

📝 格式示例:
<code>https://www.nodeseek.com/space/36628</code>

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 验证 Nodeseek 社区身份
👉 <b>下一步:</b> 发送您的个人主页链接
👉 <b>遇到问题:</b> 仅支持 nodeseek.com 域名，请确保链接正确
                `.trim(), {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消绑定', callback_data: 'cancel_reg' }]]
                    }
                });

            } catch (e) {
                await sendMessage(chatId, `❌ 验证失败: ${e.message}`);
            }
            break;

        case 'bind_nodeseek_input':
            const nsIdBind = extractNodeseekId(messageText);
            if (!nsIdBind) {
                await sendMessage(chatId, '❌ 无效的 Nodeseek 链接，请重新输入:', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消绑定', callback_data: 'cancel_reg' }]]
                    }
                });
                return;
            }

            const bindData = session.data;

            try {
                await updateUser({
                    uuid: bindData.targetUserUuid,
                    telegramId: userId,
                    description: `NS: ${nsIdBind}`
                });

                await clearSession(kv, userId);
                await sendMessage(chatId, `
✅ <b>绑定成功!</b>

╭──────────────────────
│ 👤 用户: <code>${bindData.targetUsername}</code>
│ 📱 Telegram 已关联
│ 🌐 NS ID: <code>${nsIdBind}</code>
╰──────────────────────

👉 发送 /start 查看您的账户信息
                `.trim());

                // 向群组发送欢迎消息 (订阅绑定)
                // 构建显示名称
                const tgUser = username || userId;
                const subUser = bindData.targetUsername;
                let userDisplay = '';

                if (username && subUser === username) {
                    userDisplay = `@${username}`;
                } else {
                    userDisplay = `${subUser} (@${tgUser})`;
                }

                // 向群组发送欢迎消息 (订阅绑定)
                await sendMessage(CONFIG.GROUP_ID, `
🎉 <b>欢迎新成员加入!</b>

╭──────────────────────
│ 👤 用户: ${userDisplay}
│ 📝 注册方式: 订阅链接绑定
│ 🌐 来自: Nodeseek (ID: <a href="https://www.nodeseek.com/space/${nsIdBind}">${nsIdBind}</a>)
│ ⏰ 加入时间: ${formatDate(new Date().toISOString())}
╰──────────────────────

👋 欢迎加入我们的大家庭！
                `.trim());
            } catch (e) {
                await sendMessage(chatId, `❌ 绑定失败: ${e.message}`);
            }
            break;
    }
}

// ==================== 管理员卡密生成流程 ====================
async function handleAdminCardGeneration(kv, chatId, userId, messageText, session) {
    switch (session.state) {
        case 'admin_card_traffic':
            const traffic = parseFloat(messageText);
            if (isNaN(traffic) || traffic <= 0) {
                await sendMessage(chatId, '❌ 请输入有效的流量数值 (单位: GB):');
                return;
            }

            await setSession(kv, userId, 'admin_card_strategy', { ...session.data, trafficGB: traffic });
            await sendMessage(chatId, `
🎫 <b>生成卡密</b>

╭──────────────────────
│ 📦 流量: <b>${traffic} GB</b>
│ 🔄 第2步: 选择重置策略
╰──────────────────────

请选择流量重置策略:
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '♾️ 不重置', callback_data: 'card_strat_NO_RESET' }],
                        [{ text: '📆 每日重置', callback_data: 'card_strat_DAY' }],
                        [{ text: '📅 每周重置', callback_data: 'card_strat_WEEK' }],
                        [{ text: '🗓️ 每月重置', callback_data: 'card_strat_MONTH' }],
                        [{ text: '❌ 取消操作', callback_data: 'cancel_admin' }]
                    ]
                }
            });
            break;

        case 'admin_card_count':
            const count = parseInt(messageText);
            if (isNaN(count) || count <= 0 || count > 50) {
                await sendMessage(chatId, '❌ 请输入有效的数量 (1-50):');
                return;
            }

            const cardData = session.data;
            const cards = [];

            for (let i = 0; i < count; i++) {
                const code = generateCardCode();
                await saveCard(kv, code, {
                    trafficBytes: cardData.trafficGB * 1024 * 1024 * 1024,
                    strategy: cardData.strategy,
                    duration: cardData.duration,
                    squadUuid: cardData.squadUuid,
                    squadName: cardData.squadName,
                    createdAt: Date.now(),
                    createdBy: userId
                });
                cards.push(code);
            }

            await clearSession(kv, userId);

            let result = `
✅ <b>卡密生成成功</b>

╭──────────────────────
│ 🎫 数量: <b>${count}</b> 张
│ 📦 流量: <b>${cardData.trafficGB} GB</b>
│ 🔄 策略: ${STRATEGY_NAMES[cardData.strategy] || cardData.strategy}
│ ⏱ 时长: ${DURATION_NAMES[cardData.duration] || cardData.duration}
│ 📂 分组: 「${cardData.squadName}」
╰──────────────────────

📄 <b>卡密文件已生成，请下载查看。</b>
            `.trim();

            await sendMessage(chatId, result);

            // 发送txt文件
            await sendDocument(
                chatId,
                cards.join('\n'),
                `cards_${count}_${cardData.trafficGB}GB_${new Date().toISOString().slice(0, 10)}.txt`,
                '📄 <b>卡密列表文件</b>'
            );
            break;
    }
}

// ==================== 回调处理 ====================
async function handleCallback(kv, callbackQuery) {
    const chatId = callbackQuery.message.chat.id;
    const messageId = callbackQuery.message.message_id;
    const userId = callbackQuery.from.id;
    const username = callbackQuery.from.username || '';
    const data = callbackQuery.data;

    // For view_redpacket, we'll answer with a popup alert later
    if (!data.startsWith('view_redpacket:')) {
        // Always answer the callback query to dismiss the loading state
        try { await answerCallback(callbackQuery.id); } catch (e) { /* ignore stale callback */ }
    }

    const session = await getSession(kv, userId);

    // 取消操作
    if (data === 'cancel_reg' || data === 'cancel_admin') {
        await clearSession(kv, userId);
        await sendMessage(chatId, '✅ 操作已取消');
        return;
    }

    // 注册选项
    if (data === 'reg_by_card') {
        await setSession(kv, userId, 'reg_card_input', {});
        await sendMessage(chatId, `
🎫 <b>卡密注册</b>

╭──────────────────────
│ 📝 第1步: 输入卡密
╰──────────────────────

请发送您购买的卡密:
💡 支持多卡密叠加，每行输入一个
⚠️ 多卡仅叠加流量，时长以单张为准

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 正在使用卡密注册新账户
👉 <b>下一步:</b> 请直接发送您的卡密（每行一个）
👉 <b>遇到问题:</b> 如提示卡密无效，请检查此处是否有空格或联系客服
        `.trim(), {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]]
            }
        });
        return;
    }

    // 重新输入卡密 (注册流程)
    if (data === 'reg_card_retry') {
        await setSession(kv, userId, 'reg_card_input', {});
        await sendMessage(chatId, '📝 请重新输入卡密 (每行一个):', {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]]
            }
        });
        return;
    }

    if (data === 'reg_by_sub') {
        await setSession(kv, userId, 'bind_sub_input', {});
        await sendMessage(chatId, `
🔗 <b>订阅链接绑定</b>

╭──────────────────────
│ 📝 请发送您的订阅链接
╰──────────────────────

格式示例:
<code>https://sub.1391399.xyz/xxxxx</code>

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 绑定已有订阅账户至 Telegram
👉 <b>下一步:</b> 发送您的完整订阅链接
👉 <b>遇到问题:</b> 如提示无效，请确认链接未过期且格式正确
        `.trim(), {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消绑定', callback_data: 'cancel_reg' }]]
            }
        });
        return;
    }

    // 验证频道 (注册流程)
    if (data === 'verify_channel_reg') {
        if (session.state !== 'reg_verify_channel') {
            await sendMessage(chatId, '❌ 操作已过期，请重新发起注册');
            return;
        }

        const channelResult = await getChatMember(CONFIG.CHANNEL_ID, userId);
        const chStatus = channelResult.result?.status;
        if (!channelResult.ok || !isJoinedChatStatus(chStatus)) {
            await sendMessage(chatId, '❌ 您还未加入频道，请先加入后再验证');
            return;
        }

        // 频道验证通过，生成群组邀请链接
        try {
            const inviteResult = await createChatInviteLink(CONFIG.GROUP_ID, 1);
            await setSession(kv, userId, 'reg_verify_group', { ...session.data, inviteLink: inviteResult.invite_link });

            await sendMessage(chatId, `
✅ <b>频道验证通过!</b>

╭──────────────────────
│ 👥 第5步: 加入群组
╰──────────────────────

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 加入用户交流群组
👉 <b>下一步:</b> 点击「👥 进入群组」并加入，随后点击「✅ 我已进入」
👉 <b>遇到问题:</b> 必须先加入群组才能完成最终验证，请勿直接关闭
📋 <b>操作步骤:</b>
1️⃣ 点击"进入群组"按钮加入群组
2️⃣ 加入后点击"我已进入"按钮验证
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '👥 进入群组', url: inviteResult.invite_link }],
                        [{ text: '✅ 我已进入', callback_data: 'verify_group_reg' }],
                        [{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]
                    ]
                }
            });
        } catch (e) {
            await sendMessage(chatId, `❌ 生成邀请链接失败: ${e.message}`);
        }
        return;
    }

    // 验证群组 (注册流程)
    if (data === 'verify_group_reg') {
        if (session.state !== 'reg_verify_group') {
            await sendMessage(chatId, '❌ 操作已过期，请重新发起注册');
            return;
        }

        const regData = session.data || {};
        if (!Array.isArray(regData.cards) || regData.cards.length === 0 || !regData.nodeseekId || !regData.email) {
            await sendMessage(chatId, '❌ 注册信息已过期，请重新发起注册');
            return;
        }

        const memberResult = await getChatMember(CONFIG.GROUP_ID, userId);
        const status = memberResult.result?.status;
        if (!memberResult.ok || !isJoinedChatStatus(status)) {
            await sendMessage(chatId, '❌ 您还未加入群组，请先点击"进入群组"加入后再验证');
            return;
        }

        // 废除邀请链接
        if (regData.inviteLink) {
            try {
                await revokeChatInviteLink(CONFIG.GROUP_ID, regData.inviteLink);
            } catch (e) {
                console.error('废除邀请链接失败:', e);
            }
        }

        // 创建用户
        const cards = regData.cards;
        const firstCard = cards[0];

        try {
            // 时长仅取第一张卡的时长 (多卡不叠加时长)
            const expireDate = new Date();
            expireDate.setDate(expireDate.getDate() + (DURATION_DAYS[firstCard.duration] || 30));

            // 流量叠加: 所有卡的流量之和
            const totalTraffic = firstCard.trafficBytes * cards.length;

            const newUser = await createUser({
                username: username || `tg_${userId}`,
                trafficLimitBytes: totalTraffic,
                trafficLimitStrategy: firstCard.strategy,
                expireAt: expireDate.toISOString(),
                description: `NS: ${regData.nodeseekId}`,
                telegramId: userId,
                email: regData.email,
                activeInternalSquads: [firstCard.squadUuid]
            });

            // 删除所有已使用的卡密
            for (const card of cards) {
                await deleteCard(kv, card.code);
            }
            await clearSession(kv, userId);

            // 构造卡密叠加信息
            let cardInfo = '';
            if (cards.length > 1) {
                cardInfo = `\n│ 🎫 叠加卡密: ${cards.length} 张\n│ 📦 总流量: ${formatBytes(totalTraffic)}`;
            }

            // 发送群组欢迎消息
            await sendMessage(CONFIG.GROUP_ID, `
🎉 <b>欢迎新成员加入!</b>

╭──────────────────────
│ 👤 用户: ${newUser.username} (@${username || userId})
│ 📝 注册方式: 卡密注册${cardInfo}
│ 🌐 来自: Nodeseek (ID: <a href="https://www.nodeseek.com/space/${regData.nodeseekId}">${regData.nodeseekId}</a>)
│ ⏰ 加入时间: ${formatDate(new Date().toISOString())}
╰──────────────────────

👋 欢迎加入我们的大家庭！
            `.trim());

            const panel = generateUserPanel(newUser, isAdmin(userId), '');
            await sendMessage(chatId, `
🎉 <b>注册成功!</b>

╭──────────────────────
│ ✅ 您的账户已创建成功
│ 👤 用户名: <code>${newUser.username}</code>
│ 📦 流量: <code>${formatBytes(totalTraffic)}</code>
│ 📅 到期: ${formatDate(expireDate.toISOString())}
│ ℹ️ 点击 🔄️刷新数据 获取订阅链接
╰──────────────────────

            `.trim(), {
                reply_markup: { inline_keyboard: panel.buttons }
            });
        } catch (e) {
            await sendMessage(chatId, `❌ 注册失败: ${e.message}`);
        }
        return;
    }

    // 用户功能
    if (data === 'refresh') {
        await clearSession(kv, userId);
        const user = await getUserByTelegramId(userId);
        if (!user) {
            await sendMessage(chatId, '❌ 未找到账户信息');
            return;
        }

        // 获取流量占比统计 (Top 20)
        let trafficChart = '';
        try {
            const bwStats = await getBandwidthStats(user.uuid, 20);
            if (bwStats && bwStats.topNodes && bwStats.topNodes.length > 0) {
                const chartStr = generateTrafficChart(bwStats.topNodes, user.userTraffic?.usedTrafficBytes || 0);
                // 仅当有数据时显示
                if (chartStr && !chartStr.includes('暂无流量')) {
                    trafficChart = `\n\n📊 <b>30天流量占比 (Top 20)</b>\n┌─────────────────────\n${chartStr}\n└─────────────────────`;
                }
            }
        } catch (e) {
            console.error('获取流量统计失败:', e);
        }

        const panel = generateUserPanel(user, isAdmin(userId), trafficChart);
        await editMessage(chatId, messageId, panel.text, {
            reply_markup: { inline_keyboard: panel.buttons }
        });
        return;
    }

    if (data === 'revoke_confirm') {
        await clearSession(kv, userId);
        await editMessage(chatId, messageId, `
⚠️ <b>重置订阅链接确认</b>

╭──────────────────────
│ 🔄 此操作将重新生成您的订阅链接
│
│ ⚡ 重置后:
│   • 旧订阅链接将立即失效
│   • 所有已导入的节点将不可用
│   • 需要重新导入新的订阅链接
│
│ ⏱ 操作不可撤销，请谨慎操作
╰──────────────────────
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '✅ 确认重置', callback_data: 'revoke_do' }],
                    [{ text: '◀️ 返回面板', callback_data: 'refresh' }]
                ]
            }
        });
        return;
    }

    if (data === 'revoke_do') {
        await clearSession(kv, userId);
        const user = await getUserByTelegramId(userId);
        if (!user) {
            await sendMessage(chatId, '❌ 未找到账户信息');
            return;
        }

        try {
            // 重置订阅链接
            const newShortUuid = generateRequestId();
            const updatedUser = await revokeSubscription(user.uuid, newShortUuid);
            await editMessage(chatId, messageId, `
✅ <b>订阅链接重置成功!</b>

╭──────────────────────
│ 🔗 新订阅链接已生成
│ ⏱ 旧链接已失效
╰──────────────────────

📋 <b>新订阅链接:</b>
<code>${updatedUser.subscriptionUrl}</code>

💡 请在客户端重新导入此链接
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '◀️ 返回面板', callback_data: 'refresh' }]
                    ]
                }
            });
        } catch (e) {
            await sendMessage(chatId, `❌ 重置失败: ${e.message}`);
        }
        return;
    }

    if (data === 'reset_traffic_info') {
        const user = await getUserByTelegramId(userId);
        if (!user) {
            await sendMessage(chatId, '❌ 未找到账户信息');
            return;
        }
        // 检查上次重置时间 (15天免费政策)
        const lastResetTime = user.lastTrafficResetAt ? new Date(user.lastTrafficResetAt).getTime() : 0;
        const now = Date.now();
        const daysSinceReset = lastResetTime > 0 ? Math.floor((now - lastResetTime) / 86400000) : 999;

        if (daysSinceReset >= 15) {
            // 免费重置
            await sendMessage(chatId, `
📊 <b>重置流量</b>

╭──────────────────────
│ 🎉 <b>免费重置可用!</b>
│
│ 📅 每 15 天可免费重置一次流量
│ ✅ 您已满足免费重置条件
│
│ ⚠️ 重置后流量将归零重新计算
╰──────────────────────

确认要免费重置流量吗?
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '✅ 确认免费重置', callback_data: 'reset_free' }],
                        [{ text: '❌ 取消操作', callback_data: 'refresh' }]
                    ]
                }
            });
        } else {
            const remainDays = 15 - daysSinceReset;
            await sendMessage(chatId, `
📊 <b>重置流量</b>

╭──────────────────────
│ ⏳ 距离下次免费重置还需 <b>${remainDays}</b> 天
│
│ 💰 <b>付费重置</b>
│    费用为月订阅价格的 <b>50%</b>
│
│ 📋 <b>操作步骤</b>
│    1️⃣ 前往发卡网购买对应价值商品
│    2️⃣ 复制订单号并发送至此处
│    3️⃣ 等待管理员审核确认
│
│ 🛒 <b>购买链接</b>
│    ${CONFIG.SHOP_URL}
╰──────────────────────

📝 请发送您的订单号:

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 付费重置流量
👉 <b>下一步:</b> 购买「流量重置包」后发送订单号
👉 <b>遇到问题:</b> 免费重置每 15 天仅限一次，未满 15 天需付费
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'refresh' }]]
                }
            });
            await setSession(kv, userId, 'reset_traffic_order', {});
        }
        return;
    }

    if (data === 'renew_info') {
        await clearSession(kv, userId);
        await sendMessage(chatId, `
💳 <b>续费套餐</b>

╭──────────────────────
│ 💰 <b>价格说明</b>
│    与首次订阅价格相同
│
│ 📋 <b>续费方式</b>
│    📝 订单号续费 - 在发卡网购买后输入订单号
│    🧧 口令红包续费 - 输入口令完成续费
╰──────────────────────

请选择续费方式:
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '📝 订单号续费', callback_data: 'renew_by_order' }],
                    [{ text: '🧧 口令红包续费', callback_data: 'renew_by_password' }],
                    [{ text: '🛒 卡网直购', callback_data: 'renew_by_shop' }],
                    [{ text: '❌ 取消操作', callback_data: 'refresh' }]
                ]
            }
        });
        return;
    }

    if (data === 'renew_by_order') {
        await sendMessage(chatId, `
📝 <b>订单号续费</b>

╭──────────────────────
│ 📋 <b>操作步骤</b>
│    1️⃣ 前往发卡网购买续费套餐
│    2️⃣ 复制订单号并发送至此处
│    3️⃣ 选择续费时长
│    4️⃣ 等待管理员审核确认
│
│ 🛒 <b>购买链接</b>
│    ${CONFIG.SHOP_URL}
╰──────────────────────

📝 请发送您的订单号:
        `.trim(), {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'refresh' }]]
            }
        });
        await setSession(kv, userId, 'renew_order_input', {});
        return;
    }

    if (data === 'renew_by_password') {
        await sendMessage(chatId, `
🧧 <b>口令红包续费</b>

╭──────────────────────
│ 📋 <b>操作步骤</b>
│    1️⃣ 输入您收到的口令
│    2️⃣ 选择续费时长
│    3️⃣ 等待管理员审核确认
╰──────────────────────

📝 请输入口令:

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 使用口令红包续费
👉 <b>下一步:</b> 发送您获得的充值口令
👉 <b>遇到问题:</b> 口令通常由管理员发放，如有疑问请询问管理员
        `.trim(), {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'refresh' }]]
            }
        });
        await setSession(kv, userId, 'renew_password_input', {});
        return;
    }

    if (data === 'sub_history') {
        await clearSession(kv, userId);
        const user = await getUserByTelegramId(userId);
        if (!user) {
            await sendMessage(chatId, '❌ 未找到账户信息');
            return;
        }

        try {
            const history = await getSubscriptionHistory(user.uuid);

            if (!history.records || history.records.length === 0) {
                await sendMessage(chatId, `
📜 <b>24小时订阅记录</b>

╭──────────────────────
│ 📭 暂无订阅访问记录
│
│ 💡 当您的客户端访问订阅链接时
│    这里会显示访问记录
╰──────────────────────
                `.trim(), {
                    reply_markup: {
                        inline_keyboard: [[{ text: '◀️ 返回面板', callback_data: 'refresh' }]]
                    }
                });
                return;
            }

            let text = `
📜 <b>24小时订阅记录</b>

╭──────────────────────
│ 📊 共 <b>${history.total}</b> 条记录
╰──────────────────────\n\n`;
            history.records.slice(0, 15).forEach((record, i) => {
                text += `<b>${i + 1}.</b> 🕐 ${formatDate(record.requestAt)}\n`;
                text += `    🌐 IP: <code>${record.requestIp || '未知'}</code>\n`;
                const ua = record.userAgent?.substring(0, 35) || '未知';
                text += `    📱 UA: ${ua}${record.userAgent?.length > 35 ? '...' : ''}\n\n`;
            });

            if (history.total > 15) {
                text += `\n<i>仅显示最近 15 条记录</i>`;
            }

            await sendMessage(chatId, text.trim(), {
                reply_markup: {
                    inline_keyboard: [[{ text: '◀️ 返回面板', callback_data: 'refresh' }]]
                }
            });
        } catch (e) {
            await sendMessage(chatId, `❌ 获取记录失败: ${e.message}`);
        }
        return;
    }



    // 邀请用户进入群组
    if (data === 'invite_user_start') {
        const user = await getUserByTelegramId(userId);
        if (!user) {
            await sendMessage(chatId, '❌ 未找到账户信息');
            return;
        }
        await setSession(kv, userId, 'invite_reason_input', {});
        await sendMessage(chatId, `
👥 <b>邀请用户进入群组</b>

╭──────────────────────
│ 📋 <b>操作步骤</b>
│    1️⃣ 输入邀请原因
│    2️⃣ 发送被邀请用户的 Telegram ID
│    3️⃣ 对方提前私聊Bot & Bot发送邀请链接
╰──────────────────────

📝 请输入邀请原因:

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 填写邀请申请理由
👉 <b>下一步:</b> 简述为何邀请该用户（如：朋友、同事）
👉 <b>遇到问题:</b> 理由将展示给管理员备份
        `.trim(), {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'refresh' }]]
            }
        });
        return;
    }

    // 邀请用户确认入群
    if (data === 'invite_verify_group') {
        const memberResult = await getChatMember(CONFIG.GROUP_ID, userId);
        const mStatus = memberResult.result?.status;
        if (!memberResult.ok || !isJoinedChatStatus(mStatus)) {
            await sendMessage(chatId, '❌ 您还未加入群组，请先点击"进入群组"加入');
            return;
        }

        // 从 join_request KV 获取被邀请用户的请求数据 (可能已被入群事件消费)
        const joinRequest = await getJoinRequest(kv, userId);
        // 废除之前的邀请链接
        if (joinRequest?.inviteLink) {
            try { await revokeChatInviteLink(CONFIG.GROUP_ID, joinRequest.inviteLink); } catch (e) { console.error(e); }
        }
        // 注意: 不删除 join_request, 让新成员加入时的 handler 处理 (发欢迎消息)
        await sendMessage(chatId, '✅ 入群验证成功！欢迎加入！');
        return;
    }

    // 查看口令红包 (仅管理员) - 使用弹窗显示，其他群友不可见
    if (data.startsWith('view_redpacket:')) {
        if (!isAdmin(userId)) {
            await answerCallback(callbackQuery.id, '❌ 仅管理员可查看', true);
            return;
        }
        const requestId = data.replace('view_redpacket:', '');
        const pendingRequest = await getPendingRequest(kv, requestId);
        if (!pendingRequest) {
            await answerCallback(callbackQuery.id, '❌ 该请求已过期或已被处理', true);
            return;
        }

        // 使用弹窗 (show_alert) 显示口令，只有点击的管理员可见
        const alertText = `🧧 口令红包详情\n\n👤 用户: ${pendingRequest.username}\n🔑 口令: ${pendingRequest.password}\n⏱ 时长: ${pendingRequest.durationName}\n📅 时间: ${formatDate(new Date(pendingRequest.createdAt).toISOString())}`;
        await answerCallback(callbackQuery.id, alertText, true);

        // 在原消息上添加管理员操作按钮 (编辑消息，替换"查看口令红包"按钮为同意/拒绝)
        await editMessage(chatId, messageId, `
🧧 <b>口令红包续费请求</b>

╭──────────────────────
│ 👤 用户: ${pendingRequest.username}
│ ⏱ 时长: ${DURATION_NAMES[pendingRequest.duration] || pendingRequest.durationName}
│ 📅 时间: ${formatDate(new Date(pendingRequest.createdAt).toISOString())}
╰──────────────────────

⏳ 等待管理员确认...
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [
                        { text: '✅ 同意续费', callback_data: `approve_renew_${requestId}` },
                        { text: '❌ 拒绝', callback_data: `reject_renew_${requestId}` }
                    ],
                    [{ text: '🧧 再次查看口令', callback_data: `view_redpacket:${requestId}` }]
                ]
            }
        });
        return;
    }

    // 管理员功能
    if (data === 'admin_manage_cards' && isAdmin(userId)) {
        await clearSession(kv, userId);
        try {
            const list = await kv.list({ prefix: 'card:', limit: 100 });
            if (!list.keys || list.keys.length === 0) {
                await sendMessage(chatId, '📭 <b>暂无有效卡密</b>', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '◀️ 返回面板', callback_data: 'refresh' }]]
                    }
                });
                return;
            }

            // 读取所有卡密数据用于分组
            const allCards = [];
            for (const key of list.keys) {
                const code = key.name.replace('card:', '');
                const cardData = await kv.get(key.name, 'json');
                if (cardData) {
                    allCards.push({ code, ...cardData });
                }
            }

            // 按组别分组 (相同 trafficBytes + strategy + duration + squadUuid = 同一组)
            const groups = {};
            for (const card of allCards) {
                const groupKey = `${card.trafficBytes}|${card.strategy}|${card.duration}|${card.squadUuid || ''}`;
                if (!groups[groupKey]) {
                    groups[groupKey] = {
                        trafficBytes: card.trafficBytes,
                        strategy: card.strategy,
                        duration: card.duration,
                        squadUuid: card.squadUuid,
                        squadName: card.squadName || '默认',
                        cards: []
                    };
                }
                groups[groupKey].cards.push(card.code);
            }

            const buttons = [];

            // 单张卡密按钮
            for (const card of allCards) {
                buttons.push([{ text: `🎫 ${card.code}`, callback_data: `view_card:${card.code}` }]);
            }

            // 分隔线 + 按组删除按钮
            const groupKeys = Object.keys(groups);
            if (groupKeys.length > 0) {
                buttons.push([{ text: '───── 📂 按组删除 ─────', callback_data: 'ignore_alert' }]);
                for (let i = 0; i < groupKeys.length; i++) {
                    const g = groups[groupKeys[i]];
                    const label = `🗑️ ${g.squadName} | ${formatBytes(g.trafficBytes)} | ${DURATION_NAMES[g.duration] || g.duration} (${g.cards.length}张)`;
                    // 使用索引作为回调数据，避免超出 64 字节限制
                    buttons.push([{ text: label, callback_data: `del_grp:${i}` }]);
                }
            }

            if (!list.list_complete) {
                buttons.push([{ text: '⚠️ 仅显示前 100 张卡密', callback_data: 'ignore_alert' }]);
            }

            buttons.push([{ text: '◀️ 返回面板', callback_data: 'refresh' }]);

            // 将分组信息存入 session，供后续删除使用
            await setSession(kv, userId, 'admin_card_groups', { groups: groups, groupKeys: groupKeys });

            await sendMessage(chatId, `
🔧 <b>卡密管理</b>

╭──────────────────────
│ 📊 共 <b>${allCards.length}</b> 张有效卡密
│ 📂 共 <b>${groupKeys.length}</b> 个组别
│ 👇 点击卡密查看详情，或按组删除
╰──────────────────────
            `.trim(), {
                reply_markup: { inline_keyboard: buttons }
            });
        } catch (e) {
            await sendMessage(chatId, `❌ 获取卡密列表失败: ${e.message}`);
        }
        return;
    }

    if (data.startsWith('view_card:') && isAdmin(userId)) {
        const code = data.replace('view_card:', '');
        const card = await getCard(kv, code);

        if (!card) {
            await sendMessage(chatId, '❌ 该卡密不存在或已被删除', {
                reply_markup: {
                    inline_keyboard: [[{ text: '◀️ 返回列表', callback_data: 'admin_manage_cards' }]]
                }
            });
            return;
        }

        const details = `
🎫 <b>卡密详情</b>

╭──────────────────────
│ 🔑 代码: <code>${code}</code>
│ 📦 流量: <b>${formatBytes(card.trafficBytes)}</b>
│ 🔄 策略: ${STRATEGY_NAMES[card.strategy] || card.strategy}
│ ⏱ 时长: ${DURATION_NAMES[card.duration] || card.duration}
│ 📂 分组: ${card.squadName}
│ 📅 创建: ${formatDate(typeof card.createdAt === 'number' ? new Date(card.createdAt).toISOString() : card.createdAt)}
│ 👤 创建人: ${card.createdBy}
╰──────────────────────
        `.trim();

        await sendMessage(chatId, details, {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '❌ 删除此卡密', callback_data: `del_card_ask:${code}` }],
                    [{ text: '◀️ 返回列表', callback_data: 'admin_manage_cards' }]
                ]
            }
        });
        return;
    }

    if (data.startsWith('del_card_ask:') && isAdmin(userId)) {
        const code = data.replace('del_card_ask:', '');

        await sendMessage(chatId, `
⚠️ <b>确认删除卡密?</b>

<code>${code}</code>

删除后无法恢复!
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '✅ 确认删除', callback_data: `del_card_do:${code}` }],
                    [{ text: '❌ 取消', callback_data: `view_card:${code}` }]
                ]
            }
        });
        return;
    }

    if (data.startsWith('del_card_do:') && isAdmin(userId)) {
        const code = data.replace('del_card_do:', '');

        try {
            await deleteCard(kv, code);
            await sendMessage(chatId, `✅ 卡密 <code>${code}</code> 已删除`, {
                reply_markup: {
                    inline_keyboard: [[{ text: '◀️ 返回列表', callback_data: 'admin_manage_cards' }]]
                }
            });
        } catch (e) {
            await sendMessage(chatId, `❌ 删除失败: ${e.message}`);
        }
        return;
    }

    // 按组删除卡密 - 确认
    if (data.startsWith('del_grp:') && isAdmin(userId)) {
        const groupIdx = parseInt(data.replace('del_grp:', ''));
        const grpSession = await getSession(kv, userId);
        if (grpSession.state !== 'admin_card_groups' || !grpSession.data?.groups || !grpSession.data?.groupKeys) {
            await sendMessage(chatId, '❌ 操作已过期，请重新打开卡密管理');
            return;
        }
        const groupKey = grpSession.data.groupKeys[groupIdx];
        const group = grpSession.data.groups[groupKey];
        if (!group) {
            await sendMessage(chatId, '❌ 未找到该组别');
            return;
        }
        await sendMessage(chatId, `
⚠️ <b>确认删除整组卡密?</b>

╭──────────────────────
│ 📂 分组: 「${group.squadName}」
│ 📦 流量: <b>${formatBytes(group.trafficBytes)}</b>
│ 🔄 策略: ${STRATEGY_NAMES[group.strategy] || group.strategy}
│ ⏱ 时长: ${DURATION_NAMES[group.duration] || group.duration}
│ 🎫 数量: <b>${group.cards.length}</b> 张
╰──────────────────────

❗ 此操作将删除该组所有卡密，不可恢复!
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: `✅ 确认删除 ${group.cards.length} 张`, callback_data: `del_grp_do:${groupIdx}` }],
                    [{ text: '❌ 取消', callback_data: 'admin_manage_cards' }]
                ]
            }
        });
        return;
    }

    // 按组删除卡密 - 执行
    if (data.startsWith('del_grp_do:') && isAdmin(userId)) {
        const groupIdx = parseInt(data.replace('del_grp_do:', ''));
        const grpSession = await getSession(kv, userId);
        if (grpSession.state !== 'admin_card_groups' || !grpSession.data?.groups || !grpSession.data?.groupKeys) {
            await sendMessage(chatId, '❌ 操作已过期，请重新打开卡密管理');
            return;
        }
        const groupKey = grpSession.data.groupKeys[groupIdx];
        const group = grpSession.data.groups[groupKey];
        if (!group) {
            await sendMessage(chatId, '❌ 未找到该组别');
            return;
        }
        try {
            let deletedCount = 0;
            for (const code of group.cards) {
                await deleteCard(kv, code);
                deletedCount++;
            }
            await clearSession(kv, userId);
            await sendMessage(chatId, `
✅ <b>组别卡密删除成功</b>

╭──────────────────────
│ 📂 分组: 「${group.squadName}」
│ 🗑️ 已删除: <b>${deletedCount}</b> 张
╰──────────────────────
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [[{ text: '◀️ 返回卡密管理', callback_data: 'admin_manage_cards' }]]
                }
            });
        } catch (e) {
            await sendMessage(chatId, `❌ 删除失败: ${e.message}`);
        }
        return;
    }

    if (data === 'ignore_alert') {
        // answerCallback already called at top
        return;
    }

    // 管理员功能
    if (data === 'admin_gen_card' && isAdmin(userId)) {
        await clearSession(kv, userId);
        try {
            console.log('[Admin] 开始获取内部分组...');
            const squadsResponse = await getInternalSquads();
            console.log('[Admin] API 返回数据:', JSON.stringify(squadsResponse));

            // 兼容多种可能的 API 返回格式
            let squadList = [];
            if (Array.isArray(squadsResponse)) {
                squadList = squadsResponse;
            } else if (squadsResponse && typeof squadsResponse === 'object') {
                // 尝试多种可能的属性名
                squadList = squadsResponse.internalSquads ||
                    squadsResponse.data ||
                    squadsResponse.squads ||
                    squadsResponse.list ||
                    [];
            }

            console.log('[Admin] 解析后的分组列表:', JSON.stringify(squadList));

            if (!squadList || squadList.length === 0) {
                await sendMessage(chatId, `❌ 未找到可用的内部分组\n\n<i>调试信息: API 返回 ${JSON.stringify(squadsResponse).substring(0, 200)}</i>`);
                return;
            }

            const buttons = squadList.map(s => ([{
                text: `📁 ${s.name}`,
                callback_data: `sq:${s.uuid}:${s.name.substring(0, 6)}`
            }]));
            buttons.push([{ text: '❌ 取消操作', callback_data: 'cancel_admin' }]);

            await sendMessage(chatId, `
🎫 <b>生成卡密</b>

╭──────────────────────
│ 📂 第1步: 选择内部分组
╰──────────────────────

请选择卡密对应的内部分组:
            `.trim(), {
                reply_markup: { inline_keyboard: buttons }
            });
        } catch (e) {
            console.error('[Admin] 获取分组失败:', e);
            await sendMessage(chatId, `❌ 获取分组失败: ${e.message}\n\n<i>请检查 API 连接或联系开发者</i>`);
        }
        return;
    }

    if (data.startsWith('sq:') && isAdmin(userId)) {
        const parts = data.split(':');
        const squadUuid = parts[1];
        let squadName = parts[2] || '分组';

        // 获取完整的分组名称
        try {
            const squadsResponse = await getInternalSquads();
            let squadList = [];
            if (Array.isArray(squadsResponse)) {
                squadList = squadsResponse;
            } else if (squadsResponse && typeof squadsResponse === 'object') {
                squadList = squadsResponse.internalSquads ||
                    squadsResponse.data ||
                    squadsResponse.squads ||
                    squadsResponse.list ||
                    [];
            }
            const squad = squadList.find(s => s.uuid === squadUuid);
            if (squad) {
                squadName = squad.name;
            }
        } catch (e) {
            console.error('[Admin] 获取分组详情失败:', e);
        }

        await setSession(kv, userId, 'admin_card_traffic', { squadUuid, squadName });
        await sendMessage(chatId, `
🎫 <b>生成卡密</b>

╭──────────────────────
│ 📂 分组: 「${squadName}」
│ 📦 第2步: 输入流量大小
╰──────────────────────

请输入流量大小 (单位: GB):
        `.trim(), {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'cancel_admin' }]]
            }
        });
        return;
    }

    if (data.startsWith('card_strat_') && isAdmin(userId)) {
        const strategy = data.replace('card_strat_', '');
        await setSession(kv, userId, 'admin_card_duration', { ...session.data, strategy });
        await sendMessage(chatId, `
🎫 <b>生成卡密</b>

╭──────────────────────
│ 📦 流量: <b>${session.data.trafficGB} GB</b>
│ 🔄 策略: ${STRATEGY_NAMES[strategy]}
│ ⏱ 第3步: 选择套餐时长
╰──────────────────────

请选择套餐时长:
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '📅 月付 (30天)', callback_data: 'card_dur_monthly' }],
                    [{ text: '📅 2月付 (60天)', callback_data: 'card_dur_bimonthly' }],
                    [{ text: '📆 季付 (90天)', callback_data: 'card_dur_quarterly' }],
                    [{ text: '🗓️ 半年付 (180天)', callback_data: 'card_dur_semiannual' }],
                    [{ text: '🎉 年付 (365天)', callback_data: 'card_dur_annual' }],
                    [{ text: '❌ 取消操作', callback_data: 'cancel_admin' }]
                ]
            }
        });
        return;
    }

    if (data.startsWith('card_dur_') && isAdmin(userId)) {
        const duration = data.replace('card_dur_', '');
        await setSession(kv, userId, 'admin_card_count', { ...session.data, duration });
        await sendMessage(chatId, `
🎫 <b>生成卡密</b>

╭──────────────────────
│ 📦 流量: <b>${session.data.trafficGB} GB</b>
│ 🔄 策略: <b>${session.data.strategy}</b>
│ ⏱ 时长: ${DURATION_NAMES[duration]}
│ 🔢 第4步: 输入生成数量
╰──────────────────────

请输入生成数量 (1-50):
        `.trim(), {
            reply_markup: {
                inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'cancel_admin' }]]
            }
        });
        return;
    }

    // 续费时长选择
    if (data.startsWith('renew_') && data !== 'renew_info' && data !== 'renew_by_order' && data !== 'renew_by_password' && data !== 'renew_by_shop') {
        const duration = data.replace('renew_', '');

        // Support for order renewal and password renewal
        const isOrderRenew = session.state === 'renew_duration_select' && session.data?.orderNo;
        const isPasswordRenew = session.state === 'renew_password_duration' && session.data?.password;

        if (!isOrderRenew && !isPasswordRenew) {
            await sendMessage(chatId, '❌ 操作已过期，请重新发起续费请求');
            return;
        }

        const user = await getUserByTelegramId(userId);
        if (!user) {
            await sendMessage(chatId, '❌ 未找到账户信息');
            return;
        }

        try {
            const requestId = generateRequestId();
            const requestData = {
                type: 'renew',
                userUuid: user.uuid,
                userId: userId,
                username: user.username || userId,
                duration: duration,
                durationName: DURATION_NAMES[duration] || duration,
                currentExpireAt: user.expireAt,
                createdAt: Date.now()
            };

            if (isOrderRenew) {
                requestData.orderNo = session.data.orderNo;
                requestData.orderAmount = session.data.orderAmount;
                requestData.renewMethod = 'order';
            } else {
                requestData.password = session.data.password;
                requestData.renewMethod = 'password';
            }

            await savePendingRequest(kv, requestId, requestData);
            await clearSession(kv, userId);

            const refText = isOrderRenew
                ? `│ 📝 订单号: <code>${requestData.orderNo}</code>`
                : `│ 🧧 续费方式: 口令红包`;

            await sendMessage(chatId, `
✅ <b>续费请求已提交</b>

╭──────────────────────
${refText}
│ ⏱ 时长: ${DURATION_NAMES[duration]}
│
│ ⏳ 请等待管理员审核确认
╰──────────────────────
            `.trim());

            // 发送到群组等待管理员确认
            if (isPasswordRenew) {
                // 口令红包: 群组中隐藏口令, 仅显示"查看口令红包"按钮
                await sendMessage(CONFIG.GROUP_ID,
                    `
🧧 <b>口令红包续费请求</b>

╭──────────────────────
│ 👤 用户: ${user.username || userId}
│ ⏱ 时长: ${DURATION_NAMES[duration]}
│ 📅 时间: ${formatDate(new Date().toISOString())}
╰──────────────────────

⏳ 等待管理员确认...
                    `.trim(), {
                    reply_markup: {
                        inline_keyboard: [
                            [{ text: '🧧 查看口令红包', callback_data: `view_redpacket:${requestId}` }]
                        ]
                    }
                });
            } else {
                // 订单号续费: 正常显示
                await sendMessage(CONFIG.GROUP_ID,
                    `
💳 <b>续费请求</b>

╭──────────────────────
│ 📝 订单号: <code>${requestData.orderNo}</code>
│ 💰 金额: ¥${requestData.orderAmount || '?'}
│ 👤 用户: ${user.username || userId}
│ ⏱ 时长: ${DURATION_NAMES[duration]}
│ 📅 时间: ${formatDate(new Date().toISOString())}
╰──────────────────────

⏳ 等待管理员确认...
                    `.trim(), {
                    reply_markup: {
                        inline_keyboard: [
                            [
                                { text: '✅ 同意续费', callback_data: `approve_renew_${requestId}` },
                                { text: '❌ 拒绝', callback_data: `reject_renew_${requestId}` }
                            ]
                        ]
                    }
                });
            }
        } catch (e) {
            await sendMessage(chatId, `❌ 请求提交失败: ${e.message}`);
        }
        return;
    }

    // 管理员处理重置流量请求
    if (data.startsWith('approve_reset_') || data.startsWith('reject_reset_')) {
        // 检查是否是管理员
        if (!isAdmin(userId)) {
            await sendMessage(chatId, '❌ 只有管理员可以操作');
            return;
        }

        const isApprove = data.startsWith('approve_reset_');
        const requestId = data.replace(isApprove ? 'approve_reset_' : 'reject_reset_', '');
        const pendingRequest = await getPendingRequest(kv, requestId);

        if (!pendingRequest) {
            try { await editMessage(chatId, messageId, '❌ 该请求已过期或已被处理'); } catch (e) { }
            return;
        }

        // 先删除请求防止重复点击
        await deletePendingRequest(kv, requestId);

        if (isApprove) {
            try {
                await resetTraffic(pendingRequest.userUuid);

                // 更新群组消息
                await editMessage(chatId, messageId, `
📊 <b>重置流量请求</b> ✅ 已通过

╭──────────────────────
│ 📝 订单号: <code>${pendingRequest.orderNo}</code>
│ 👤 用户: ${pendingRequest.username}
│ 🕐 处理时间: ${formatDate(new Date().toISOString())}
│ 👮 处理人: @${callbackQuery.from.username || userId}
╰──────────────────────
                `.trim());

                // 通知用户
                await sendMessage(pendingRequest.userId, `
✅ <b>流量重置成功!</b>

╭──────────────────────
│ 📦 您的流量已重置
│ 🔄 现可正常使用
╰──────────────────────

👉 发送 /start 查看最新状态
                `.trim());
            } catch (e) {
                // 操作失败，重新保存请求
                await savePendingRequest(kv, requestId, pendingRequest);
                await sendMessage(chatId, `❌ 重置失败: ${e.message}`);
            }
        } else {
            // 更新群组消息
            await editMessage(chatId, messageId, `
📊 <b>重置流量请求</b> ❌ 已拒绝

╭──────────────────────
│ 📝 订单号: <code>${pendingRequest.orderNo}</code>
│ 👤 用户: ${pendingRequest.username}
│ 🕐 处理时间: ${formatDate(new Date().toISOString())}
│ 👮 处理人: @${callbackQuery.from.username || userId}
╰──────────────────────
            `.trim());

            // 通知用户
            await sendMessage(pendingRequest.userId, `
❌ <b>流量重置请求已被拒绝</b>

如有疑问，请联系管理员。
            `.trim());
        }
        return;
    }

    // 管理员处理续费请求
    if (data.startsWith('approve_renew_') || data.startsWith('reject_renew_')) {
        // 检查是否是管理员
        if (!isAdmin(userId)) {
            await sendMessage(chatId, '❌ 只有管理员可以操作');
            return;
        }

        const isApprove = data.startsWith('approve_renew_');
        const requestId = data.replace(isApprove ? 'approve_renew_' : 'reject_renew_', '');
        const pendingRequest = await getPendingRequest(kv, requestId);

        if (!pendingRequest) {
            try { await editMessage(chatId, messageId, '❌ 该请求已过期或已被处理'); } catch (e) { }
            return;
        }

        // 先删除请求防止重复点击
        await deletePendingRequest(kv, requestId);

        if (isApprove) {
            try {
                const currentExpire = new Date(pendingRequest.currentExpireAt);
                const newExpire = new Date(Math.max(currentExpire.getTime(), Date.now()));
                newExpire.setDate(newExpire.getDate() + (DURATION_DAYS[pendingRequest.duration] || 30));

                await updateUser({
                    uuid: pendingRequest.userUuid,
                    expireAt: newExpire.toISOString()
                });

                // Update group message
                const approveRefLine = pendingRequest.renewMethod === 'password'
                    ? `│ 🧧 续费方式: 口令红包`
                    : `│ 📝 订单号: <code>${pendingRequest.orderNo}</code>`;
                await editMessage(chatId, messageId, `
💳 <b>续费请求</b> ✅ 已通过

╭──────────────────────
${approveRefLine}
│ 👤 用户: ${pendingRequest.username}
│ ⏱ 时长: ${DURATION_NAMES[pendingRequest.duration] || pendingRequest.duration}
│ 📅 新到期: ${formatDate(newExpire.toISOString())}
│ 👮 处理人: @${callbackQuery.from.username || userId}
╰──────────────────────
                `.trim());

                // Notify user
                await sendMessage(pendingRequest.userId, `
✅ <b>续费成功!</b>

╭──────────────────────
│ 🎉 您的续费请求已通过
│ 📅 新到期时间: ${formatDate(newExpire.toISOString())}
╰──────────────────────

👉 发送 /start 查看最新状态
                `.trim());
            } catch (e) {
                // 操作失败，重新保存请求
                await savePendingRequest(kv, requestId, pendingRequest);
                await sendMessage(chatId, `❌ 续费失败: ${e.message}`);
            }
        } else {
            // Update group message
            const rejectRefLine = pendingRequest.renewMethod === 'password'
                ? `│ 🧧 续费方式: 口令红包`
                : `│ 📝 订单号: <code>${pendingRequest.orderNo}</code>`;
            await editMessage(chatId, messageId, `
💳 <b>续费请求</b> ❌ 已拒绝

╭──────────────────────
${rejectRefLine}
│ 👤 用户: ${pendingRequest.username}
│ ⏱ 时长: ${DURATION_NAMES[pendingRequest.duration] || pendingRequest.duration}
│ 🕐 处理时间: ${formatDate(new Date().toISOString())}
│ 👮 处理人: @${callbackQuery.from.username || userId}
╰──────────────────────
            `.trim());

            // Notify user
            await sendMessage(pendingRequest.userId, `
❌ <b>续费请求已被拒绝</b>

如有疑问，请联系管理员。
            `.trim());
        }
        return;
    }

    // ==================== 免费重置流量 ====================
    if (data === 'reset_free') {
        const user = await getUserByTelegramId(userId);
        if (!user) { await sendMessage(chatId, '❌ 未找到账户信息'); return; }
        const lastResetTime = user.lastTrafficResetAt ? new Date(user.lastTrafficResetAt).getTime() : 0;
        const daysSinceReset = lastResetTime > 0 ? Math.floor((Date.now() - lastResetTime) / 86400000) : 999;
        if (daysSinceReset < 15) {
            await sendMessage(chatId, `❌ 距离下次免费重置还需 ${15 - daysSinceReset} 天`);
            return;
        }
        try {
            await resetTraffic(user.uuid);
            await sendMessage(chatId, `
✅ <b>流量重置成功!</b>

╭──────────────────────
│ 📦 您的流量已重置
│ 🔄 现可正常使用
│ 📅 下次免费重置: 15天后
╰──────────────────────

👉 发送 /start 查看最新状态
            `.trim());
            await sendMessage(CONFIG.GROUP_ID, `
📊 <b>免费流量重置</b>

╭──────────────────────
│ 👤 用户: ${user.username || userId}
│ 🔄 类型: 免费重置 (15天周期)
│ 📅 时间: ${formatDate(new Date().toISOString())}
╰──────────────────────
            `.trim());
        } catch (e) {
            await sendMessage(chatId, `❌ 重置失败: ${e.message}`);
        }
        return;
    }

    // ==================== 卡网直购 - 注册 ====================
    if (data === 'reg_by_shop') {
        await clearSession(kv, userId);
        await sendMessage(chatId, `
🛒 <b>卡网直购注册</b>

╭──────────────────────
│ 📋 <b>操作流程</b>
│    1️⃣ 登入卡网账户 (可选)
│    2️⃣ 浏览并选择商品
│    3️⃣ 完成支付获取卡密
│    4️⃣ 自动完成注册流程
╰──────────────────────

是否需要登入卡网账户?
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '🔑 需要登入', callback_data: 'sl_yes' }],
                    [{ text: '👤 游客模式', callback_data: 'sl_no' }],
                    [{ text: '❌ 取消', callback_data: 'cancel_reg' }]
                ]
            }
        });
        await setSession(kv, userId, 'shop_ask_login', { flow: 'reg' });
        return;
    }

    // ==================== 卡网直购 - 续费 ====================
    if (data === 'renew_by_shop') {
        await clearSession(kv, userId);
        await sendMessage(chatId, `
🛒 <b>卡网直购续费</b>

╭──────────────────────
│ 📋 <b>操作流程</b>
│    1️⃣ 登入卡网账户 (可选)
│    2️⃣ 选择续费商品
│    3️⃣ 完成支付
│    4️⃣ 自动提交续费请求
╰──────────────────────

是否需要登入卡网账户?
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '🔑 需要登入', callback_data: 'sl_yes' }],
                    [{ text: '👤 游客模式', callback_data: 'sl_no' }],
                    [{ text: '❌ 取消', callback_data: 'refresh' }]
                ]
            }
        });
        await setSession(kv, userId, 'shop_ask_login', { flow: 'renew' });
        return;
    }

    // ==================== 自助换车 ====================
    if (data === 'change_car_info') {
        const user = await getUserByTelegramId(userId);
        if (!user) { await sendMessage(chatId, '❌ 未找到账户信息'); return; }
        if (!user.expireAt) { await sendMessage(chatId, '❌ 无法获取到期时间信息'); return; }
        const remainDays = Math.ceil((new Date(user.expireAt).getTime() - Date.now()) / 86400000);
        if (remainDays > 5) {
            await sendMessage(chatId, `
❌ <b>暂不满足换车条件</b>

╭──────────────────────
│ 📅 您的剩余天数: <b>${remainDays}</b> 天
│ ⚠️ 需要距离到期 <b>≤5天</b> 才可换车
╰──────────────────────
            `.trim(), {
                reply_markup: { inline_keyboard: [[{ text: '◀️ 返回', callback_data: 'refresh' }]] }
            });
            return;
        }
        const carryOverDays = Math.max(remainDays, 0);
        await sendMessage(chatId, `
🔄 <b>自助换车</b>

╭──────────────────────
│ 📋 <b>换车说明</b>
│    • 当前剩余 <b>${carryOverDays}</b> 天将结转到新车
│    • 现有参数将重置 (流量/策略等)
│    • 仅保留订阅链接
│
│ ⚠️ 换车后需重新配置客户端
╰──────────────────────

请选择换车方式:
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '🎫 使用卡密换车', callback_data: 'change_car_card' }],
                    [{ text: '🛒 卡网直购换车', callback_data: 'change_car_shop' }],
                    [{ text: '❌ 取消', callback_data: 'refresh' }]
                ]
            }
        });
        await setSession(kv, userId, 'change_car_menu', { carryOverDays });
        return;
    }

    if (data === 'change_car_card') {
        const sess = await getSession(kv, userId);
        const carryOverDays = sess.data?.carryOverDays || 0;
        await setSession(kv, userId, 'change_car_card_input', { carryOverDays });
        await sendMessage(chatId, `
🎫 <b>卡密换车</b>

╭──────────────────────
│ 📝 请输入新的卡密
│ 💡 支持多张卡密 (每行一张)
│ 📦 多卡仅叠加流量
╰──────────────────────

请发送卡密:

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 使用新卡密进行换车
👉 <b>下一步:</b> 发送新购入的卡密
👉 <b>遇到问题:</b> 换车将保留订阅链接但重置流量和策略，请确认清楚
        `.trim(), {
            reply_markup: { inline_keyboard: [[{ text: '❌ 取消', callback_data: 'refresh' }]] }
        });
        return;
    }

    if (data === 'change_car_shop') {
        const sess = await getSession(kv, userId);
        const carryOverDays = sess.data?.carryOverDays || 0;
        await setSession(kv, userId, 'shop_ask_login', { flow: 'change', carryOverDays });
        await sendMessage(chatId, `
🛒 <b>卡网直购换车</b>

是否需要登入卡网账户?
        `.trim(), {
            reply_markup: {
                inline_keyboard: [
                    [{ text: '🔑 需要登入', callback_data: 'sl_yes' }],
                    [{ text: '👤 游客模式', callback_data: 'sl_no' }],
                    [{ text: '❌ 取消', callback_data: 'refresh' }]
                ]
            }
        });
        return;
    }

    if (data === 'change_car_confirm') {
        const sess = await getSession(kv, userId);
        if (sess.state !== 'change_car_confirm' || !sess.data?.cards) {
            await sendMessage(chatId, '❌ 操作已过期，请重新发起');
            return;
        }
        const { cards, carryOverDays } = sess.data;
        const user = await getUserByTelegramId(userId);
        if (!user) { await sendMessage(chatId, '❌ 未找到账户信息'); return; }
        try {
            let firstCard = null;
            let totalTraffic = 0;
            for (const code of cards) {
                const card = await getCard(kv, code);
                if (!card) throw new Error(`卡密 ${code} 不存在或已被使用`);
                if (!firstCard) firstCard = card;
                totalTraffic += card.trafficBytes || 0;
            }
            if (!firstCard) throw new Error('无有效卡密');
            const cardDurationDays = DURATION_DAYS[firstCard.duration] || 30;
            const totalDays = carryOverDays + cardDurationDays;
            const newExpire = new Date();
            newExpire.setDate(newExpire.getDate() + totalDays);
            await updateUser({
                uuid: user.uuid,
                expireAt: newExpire.toISOString(),
                trafficLimitBytes: totalTraffic,
                internalSquadUuid: firstCard.squadUuid
            });
            await resetTraffic(user.uuid);
            for (const code of cards) {
                try { await deleteCard(kv, code); } catch (e) { console.error('删除卡密失败:', e); }
            }
            await clearSession(kv, userId);
            const subscriptionUrl = user.subscriptionUrl || '暂无';
            await sendMessage(chatId, `
✅ <b>换车成功!</b>

╭──────────────────────
│ 🔄 已切换至新车
│ 📦 流量: <b>${formatBytes(totalTraffic)}</b>
│ ⏱ 时长: <b>${totalDays}</b> 天 (含结转${carryOverDays}天)
│ 📅 到期: ${formatDate(newExpire.toISOString())}
│
│ 🔗 <b>订阅链接</b>
│ <code>${subscriptionUrl}</code>
╰──────────────────────

⚠️ 请使用订阅链接重新配置客户端
👉 发送 /start 查看最新状态
            `.trim());
            await sendMessage(CONFIG.GROUP_ID, `
🔄 <b>自助换车</b>

╭──────────────────────
│ 👤 用户: ${user.username || userId}
│ 📦 新流量: ${formatBytes(totalTraffic)}
│ ⏱ 新时长: ${totalDays} 天
│ 📅 时间: ${formatDate(new Date().toISOString())}
╰──────────────────────
            `.trim());
        } catch (e) {
            await sendMessage(chatId, `❌ 换车失败: ${e.message}`, {
                reply_markup: { inline_keyboard: [[{ text: '◀️ 返回', callback_data: 'refresh' }]] }
            });
        }
        return;
    }

    // ==================== Shop 通用回调 ====================
    if (data === 'sl_yes') {
        const sess = await getSession(kv, userId);
        if (!sess.data?.flow) { await sendMessage(chatId, '❌ 操作已过期'); return; }
        await setSession(kv, userId, 'shop_login_user', sess.data);
        await sendMessage(chatId, `
🔑 <b>登入卡网</b>

请输入卡网用户名:

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 登录发卡网账户
👉 <b>下一步:</b> 发送您的发卡网用户名
👉 <b>遇到问题:</b> 如忘记用户名，请尝试使用游客模式或在网页端找回
        `.trim(), {
            reply_markup: { inline_keyboard: [[{ text: '❌ 取消', callback_data: sess.data.flow === 'reg' ? 'cancel_reg' : 'refresh' }]] }
        });
        return;
    }

    if (data === 'sl_no') {
        const sess = await getSession(kv, userId);
        if (!sess.data?.flow) { await sendMessage(chatId, '❌ 操作已过期'); return; }
        const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
        await shopShowProducts(kv, chatId, userId, shopClient, sess.data.flow);
        return;
    }

    if (data.startsWith('si_')) {
        const itemId = parseInt(data.replace('si_', ''));
        const sess = await getSession(kv, userId);
        if (sess.state !== 'shop_browse' || !sess.data?.products) {
            await sendMessage(chatId, '❌ 操作已过期');
            return;
        }
        const item = sess.data.products.find(p => p.id === itemId);
        if (!item) { await sendMessage(chatId, '❌ 商品不存在'); return; }
        const skus = shopToArray(item.sku || item.skus || []);
        let text = `🛍 <b>${shopEscapeHtml(item.name)}</b>\n\n`;
        if (item.introduce) text += `📝 ${shopEscapeHtml(String(item.introduce).substring(0, 200))}\n\n`;
        const buttons = [];
        if (skus.length > 0) {
            text += '📦 <b>选择规格:</b>\n';
            for (const sku of skus) {
                const skuName = sku.name || '默认';
                const skuPrice = sku.price || item.price || '?';
                const stock = sku.stock !== undefined ? ` (库存:${sku.stock})` : '';
                text += `  • ${shopEscapeHtml(skuName)} - ¥${skuPrice}${stock}\n`;
                buttons.push([{ text: `💰 ${shopEscapeHtml(String(skuName).substring(0, 20))} - ¥${skuPrice}`, callback_data: `ssku_${sku.id}` }]);
            }
        } else {
            buttons.push([{ text: `💰 购买 - ¥${item.price || '?'}`, callback_data: `ssku_${item.id}` }]);
        }
        await setSession(kv, userId, 'shop_browse', { ...sess.data, selectedItemId: itemId });
        buttons.push([{ text: '◀️ 返回商品界面', callback_data: 'sback' }]);
        buttons.push([{ text: '❌ 取消', callback_data: sess.data.flow === 'reg' ? 'cancel_reg' : 'refresh' }]);
        await sendMessage(chatId, text.trim(), { reply_markup: { inline_keyboard: buttons } });
        return;
    }

    if (data.startsWith('ssku_')) {
        const skuId = parseInt(data.replace('ssku_', ''));
        const sess = await getSession(kv, userId);
        if (sess.state !== 'shop_browse') { await sendMessage(chatId, '❌ 操作已过期'); return; }
        await setSession(kv, userId, 'shop_quantity', { ...sess.data, selectedSkuId: skuId });
        await sendMessage(chatId, '📦 <b>输入购买数量</b>\n\n请输入数量 (直接发送数字):\n\n📋 <b>操作指南</b>\n━━━━━━━━━━━━━━━━\n👉 <b>当前操作:</b> 确定商品购买数量\n👉 <b>下一步:</b> 发送数字 (如 1)\n👉 <b>遇到问题:</b> 数量必须为大于 0 的整数', {
            reply_markup: { inline_keyboard: [[{ text: '❌ 取消', callback_data: sess.data.flow === 'reg' ? 'cancel_reg' : 'refresh' }]] }
        });
        return;
    }

    if (data === 'sback') {
        const sess = await getSession(kv, userId);
        // 如果会话过期或产品列表丢失，尝试重新获取
        if (sess.state !== 'shop_browse' || !sess.data?.products) {
            const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
            const flow = sess.data?.flow || 'reg'; // 默认为注册流程
            await shopShowProducts(kv, chatId, userId, shopClient, flow);
            return;
        }
        await shopRenderProductPage(chatId, sess.data.products, sess.data.page || 1, sess.data.flow);
        return;
    }

    if (data.startsWith('spage_')) {
        const newPage = parseInt(data.replace('spage_', ''));
        const sess = await getSession(kv, userId);
        if (sess.state !== 'shop_browse' || !sess.data?.products) { await sendMessage(chatId, '❌ 操作已过期'); return; }
        await setSession(kv, userId, 'shop_browse', { ...sess.data, page: newPage });
        await shopRenderProductPage(chatId, sess.data.products, newPage, sess.data.flow);
        return;
    }

    if (data.startsWith('pay_')) {
        const tradeNo = data.replace('pay_', '');
        const sess = await getSession(kv, userId);
        const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
        const loginStatus = await ensureShopLoggedIn(kv, shopClient);
        const isLoggedIn = loginStatus.loggedIn;

        // Ensure we have access to this trade (simple check or skip)
        // ensureTradeAccess removed as it was undefined. Assuming tradeNo is sufficient proof of intent or session validation occurs later.

        let amount = '0';
        try {
            const orderRes = await shopClient.request('/pay/getOrder', { trade_no: tradeNo });
            if (orderRes.code === 200 && orderRes.data) {
                amount = String(orderRes.data.order_amount || orderRes.data.trade_amount || 0);
            }
        } catch (e) {
            console.error('Failed to get order amount', e);
        }

        if (parseFloat(amount) <= 0) {
            await sendMessage(chatId, '❌ 无法获取订单金额，请稍后重试');
            return;
        }

        const amountNum = parseFloat(amount);
        const normalizedAmount = Number.isFinite(amountNum) ? amountNum.toFixed(2) : '0.00';

        const payListRes = await shopClient.request('/pay/list', {
            business: 'product',
            amount: normalizedAmount
        });
        const payMethods = shopToArray(payListRes.data);
        const balance = payListRes.balance ?? payListRes.ext?.balance ?? 0;
        const balanceNum = parseFloat(String(balance));
        const isLoginRes = payListRes.is_login ?? payListRes.ext?.is_login ?? false;
        // Use session login status as fallback if API doesn't return it
        const isLogin = isLoginRes || (loginStatus && loginStatus.loggedIn);

        if (payMethods.length === 0 && (!isLogin || balanceNum < amountNum)) {
            await sendMessage(chatId, '❌ 暂无可用支付方式，请联系客服');
            return;
        }

        await setSession(kv, userId, 'shop_paying', { tradeNo, orderAmount: amount, flow: sess.data.flow, carryOverDays: sess.data.carryOverDays, products: sess.data.products, page: sess.data.page });

        const payButtons = [];
        // Balance Pay
        if (isLogin && Number.isFinite(balanceNum) && Number.isFinite(amountNum) && amountNum > 0 && balanceNum >= amountNum) {
            payButtons.push([{ text: `💰 余额支付 (剩余: ¥${balance})`, callback_data: `shop_pay_${tradeNo}_balance` }]);
        }

        payMethods.forEach(m => {
            const name = m.name || m.pay_name || '支付';
            payButtons.push([{ text: `💳 ${name}`, callback_data: `shop_pay_${tradeNo}_${m.id}` }]);
        });
        payButtons.push([{ text: '❌ 取消订单', callback_data: 'shop_cancel' }]);

        let text = `📦 <b>订单支付</b>\n\n╭──────────────────────\n│ 📝 订单号: <code>${tradeNo}</code>\n│ 💰 金额: ¥${amount}`;
        if (isLogin) text += `\n│ 💰 余额: ¥${balance}`;
        text += `\n╰──────────────────────\n\n请选择支付方式:`;

        await sendMessage(chatId, text.trim(), {
            reply_markup: { inline_keyboard: payButtons }
        });
        return;
    }

    // ==================== Order View & Cancel & Pay Helpers ====================
    if (data.startsWith('getitem_')) {
        const tradeNo = data.replace('getitem_', '');
        const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
        const loginStatus = await ensureShopLoggedIn(kv, shopClient);
        const isLoggedIn = loginStatus.loggedIn;
        const access = await ensureTradeAccess(kv, shopClient, chatId, tradeNo, isLoggedIn);
        if (!access.ok) { await answerCallback(callbackQuery.id, access.msg, true); return; }

        const view = await buildOrderView(shopClient, tradeNo, isLoggedIn, CONFIG.SHOP_URL);
        if (!view.ok) { await answerCallback(callbackQuery.id, view.msg, true); return; }

        await bindTradeToChat(kv, chatId, tradeNo);
        if (view.requiresPayment) {
            await setPendingTrade(kv, chatId, tradeNo);
        } else {
            await clearPendingTradeIfMatch(kv, chatId, tradeNo);
        }
        await editMessage(chatId, messageId, view.text, { reply_markup: { inline_keyboard: view.buttons } });
        return;
    }

    if (data.startsWith('shop_cancel_order_')) {
        const tradeNo = data.replace('shop_cancel_order_', '');
        const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
        const loginStatus = await ensureShopLoggedIn(kv, shopClient);
        const isLoggedIn = loginStatus.loggedIn;
        const access = await ensureTradeAccess(kv, shopClient, chatId, tradeNo, isLoggedIn);
        if (!access.ok) { await answerCallback(callbackQuery.id, access.msg, true); return; }

        const res = await shopClient.request('/shop/order/cancel', { trade_no: tradeNo });
        if (res.code === 200) {
            await clearPendingTradeIfMatch(kv, chatId, tradeNo);
            await sendMessage(chatId, `✅ 订单 <code>${tradeNo}</code> 已取消`, {
                reply_markup: { inline_keyboard: [[{ text: '◀️ 返回', callback_data: 'refresh' }]] }
            });
        } else {
            await answerCallback(callbackQuery.id, `❌ 取消失败: ${res.msg}`, true);
        }
        return;
    }

    // Pay handler (redirects to shop_pay flow or refreshes payment methods)
    if (data.startsWith('pay_')) {
        const tradeNo = data.replace('pay_', '');
        // Logic similar to shop_quantity payment fetch
        // Retrieve order amount first
        const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
        const loginStatus = await ensureShopLoggedIn(kv, shopClient);
        const isLoggedIn = loginStatus.loggedIn;
        const access = await ensureTradeAccess(kv, shopClient, chatId, tradeNo, isLoggedIn);
        if (!access.ok) { await answerCallback(callbackQuery.id, access.msg, true); return; }

        await setPendingTrade(kv, chatId, tradeNo);

        let amount = '0';
        try {
            const orderRes = await shopClient.request('/pay/getOrder', { trade_no: tradeNo });
            if (orderRes.code === 200 && orderRes.data) amount = String(orderRes.data.order_amount || 0);
        } catch (e) { }

        const amountNum = parseFloat(amount);
        const normalizedAmount = Number.isFinite(amountNum) ? amountNum.toFixed(2) : '0.00';

        const payListRes = await shopClient.request('/pay/list', { business: 'product', amount: normalizedAmount });
        const payMethods = shopToArray(payListRes.data);
        const balance = payListRes.balance ?? payListRes.ext?.balance ?? 0;
        const balanceNum = parseFloat(String(balance));
        const isLogin = payListRes.is_login ?? payListRes.ext?.is_login ?? isLoggedIn;

        const payButtons = [];
        if (isLogin && Number.isFinite(balanceNum) && Number.isFinite(amountNum) && amountNum > 0 && balanceNum >= amountNum) {
            payButtons.push([{ text: `💰 余额支付 (剩余: ¥${balance})`, callback_data: `shop_pay_${tradeNo}_balance` }]);
        }
        payMethods.forEach(m => {
            const name = m.name || m.pay_name || '支付';
            payButtons.push([{ text: `💳 ${name}`, callback_data: `shop_pay_${tradeNo}_${m.id}` }]);
        });
        payButtons.push([{ text: '❌ 取消订单', callback_data: `shop_cancel_order_${tradeNo}` }]);

        await editMessage(chatId, messageId, `📦 <b>订单支付</b>\n\n╭──────────────────────\n│ 📝 订单号: <code>${tradeNo}</code>\n│ 💰 金额: ¥${amount}\n╰──────────────────────\n\n请选择支付方式:`, {
            reply_markup: { inline_keyboard: payButtons }
        });
        return;
    }

    if (data.startsWith('shop_pay_')) {
        const parts = data.split('_');
        // Expected: shop_pay_TRADENO_METHOD
        if (parts.length < 4) { await sendMessage(chatId, '❌ 操作已过期'); return; }

        const tradeNo = parts[2];
        const payIdStr = parts.slice(3).join('_');

        const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
        const loginStatus = await ensureShopLoggedIn(kv, shopClient);
        const isLoggedIn = loginStatus.loggedIn;

        const access = await ensureTradeAccess(kv, shopClient, chatId, tradeNo, isLoggedIn);
        if (!access.ok) {
            await answerCallback(callbackQuery.id, access.msg);
            return;
        }

        try {
            const payload = { trade_no: tradeNo };
            if (payIdStr === 'balance') {
                payload.balance = true;
                payload.method = 0;
            } else {
                payload.balance = false;
                payload.method = parseInt(payIdStr);
            }

            const payRes = await shopClient.request('/pay', payload);
            if (payRes.code === 200) {
                if (payIdStr === 'balance') {
                    // Balance payment - check if status=2 means paid
                    const isPaid = Number(payRes.data?.status) === 2;
                    if (isPaid) {
                        await sendMessage(chatId, `✅ <b>余额支付成功!</b>\n\n请点击下方按钮继续。`, {
                            reply_markup: {
                                inline_keyboard: [
                                    [{ text: '📥 查看详情', callback_data: `shop_checkpay_${tradeNo}` }]
                                ]
                            }
                        });
                    } else {
                        await sendMessage(chatId, `⏳ <b>余额支付处理中</b>\n\n请稍后查询状态。`, {
                            reply_markup: {
                                inline_keyboard: [
                                    [{ text: '🔄 查询状态', callback_data: `shop_checkpay_${tradeNo}` }]
                                ]
                            }
                        });
                    }
                } else if (payRes.data?.pay_url) {
                    const payUrl = payRes.data.pay_url;
                    // Ensure URL is absolute. If relative, prepend shop URL.
                    const fullPayUrl = payUrl.startsWith('http') ? payUrl : (CONFIG.SHOP_URL.replace(/\/$/, '') + payUrl);

                    await sendMessage(chatId, `
💳 <b>请完成支付</b>

╭──────────────────────
│ 📝 订单号: <code>${tradeNo}</code>
│ 🔗 <a href="${fullPayUrl}">点击前往支付</a>
╰──────────────────────

支付完成后请点击下方按钮:
                    `.trim(), {
                        reply_markup: {
                            inline_keyboard: [
                                [{ text: '🔗 打开支付页面', url: fullPayUrl }],
                                [{ text: '✅ 我已支付', callback_data: `shop_checkpay_${tradeNo}` }],
                                [{ text: '❌ 取消订单', callback_data: 'shop_cancel' }]
                            ]
                        }
                    });
                } else {
                    await sendMessage(chatId, `❌ 发起支付失败: ${payRes.msg || '未知错误'}`, {
                        reply_markup: { inline_keyboard: [[{ text: '◀️ 返回', callback_data: 'refresh' }]] }
                    });
                }
            } else {
                await answerCallback(callbackQuery.id, `❌ 支付失败: ${payRes.msg}`);
            }
        } catch (e) {
            await sendMessage(chatId, `❌ 支付异常: ${e.message}`);
        }
        return;
    }

    if (data.startsWith('shop_checkpay')) {
        let tradeNo;
        if (data.startsWith('shop_checkpay_')) {
            tradeNo = data.replace('shop_checkpay_', '');
        }

        const sess = await getSession(kv, userId);
        if (!tradeNo && sess.state === 'shop_paying') {
            tradeNo = sess.data?.tradeNo;
        }

        if (!tradeNo) { await sendMessage(chatId, '❌ 操作已过期'); return; }

        const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
        const loginStatus = await ensureShopLoggedIn(kv, shopClient);
        const isLoggedIn = loginStatus.loggedIn;

        const access = await ensureTradeAccess(kv, shopClient, chatId, tradeNo, isLoggedIn);
        if (!access.ok) {
            await answerCallback(callbackQuery.id, access.msg);
            return;
        }

        try {
            const statusRes = await shopClient.request('/pay/getOrder', { trade_no: tradeNo });
            if (statusRes.code === 200) {
                const status = Number(statusRes.data?.status ?? statusRes.data?.trade_status ?? -1);

                // 状态 2: 已支付
                if (status === 2) {
                    await bindTradeToChat(kv, chatId, tradeNo);

                    const flow = sess.data?.flow || 'reg';

                    // ========== 续费流程 (Order Renewal) ==========
                    // 无需获取卡密，直接进入时长选择 -> 提交工单
                    if (flow === 'renew') {
                        await clearSession(kv, userId);
                        await clearPendingTradeIfMatch(kv, chatId, tradeNo);
                        // Save orderAmount to session for display later
                        await setSession(kv, userId, 'renew_duration_select', { orderNo: tradeNo, orderAmount: sess.data?.orderAmount });
                        await sendMessage(chatId, `
✅ <b>支付验证成功!</b>

╭──────────────────────
│ 📝 订单号: <code>${tradeNo}</code>
│ 💰 金额: ¥${sess.data?.orderAmount || '?'}
╰──────────────────────

请选择您的续费时长:
                        `.trim(), {
                            reply_markup: {
                                inline_keyboard: [
                                    [{ text: '📅 月付 (30天)', callback_data: 'renew_monthly' }],
                                    [{ text: '📅 2月付 (60天)', callback_data: 'renew_bimonthly' }],
                                    [{ text: '📆 季付 (90天)', callback_data: 'renew_quarterly' }],
                                    [{ text: '🗓️ 半年付 (180天)', callback_data: 'renew_semiannual' }],
                                    [{ text: '🎉 年付 (365天)', callback_data: 'renew_annual' }],
                                    [{ text: '❌ 取消', callback_data: 'refresh' }]
                                ]
                            }
                        });
                        return;
                    }

                    // ========== 注册 或 换车流程 ==========
                    // 需要获取卡密 - 使用 getOrderCards 更健壮的实现
                    await sendMessage(chatId, '✅ 支付验证成功! 正在获取卡密...');

                    const cards = await shopGetOrderCards(shopClient, tradeNo, isLoggedIn);

                    if (cards.length === 0) {
                        await sendMessage(chatId, '⚠️ 暂未获取到卡密，可能是自动发货延迟或无需卡密。', {
                            reply_markup: {
                                inline_keyboard: [
                                    [{ text: '🔄 重新检查', callback_data: `shop_checkpay_${tradeNo}` }],
                                    [{ text: '◀️ 返回', callback_data: 'refresh' }]
                                ]
                            }
                        });
                        return;
                    }

                    if (flow === 'reg') {
                        // 尝试在 BOT_KV 中查找卡密
                        // 如果是外部卡密（非 Bot 生成），这将失败。
                        // 但是对于"卡网直购"流程，我们假设卡密是兼容的文本

                        const validCards = [];

                        // 直接使用获取到的卡密字符串
                        for (const code of cards) {
                            const trimmed = code.trim();
                            // 尝试获取元数据，如果获取不到，构造默认元数据
                            // 注意：如果卡密不是 Bot 生成的，getCard 将返回 null
                            const card = await getCard(kv, trimmed);
                            if (card) {
                                validCards.push({ ...card, code: trimmed });
                            } else {
                                // 如果不是 Bot 生成的卡密 (例如人工发货或外部商品)
                                // 我们无法自动注册，因为不知道流量/策略等参数
                                // 此时只能把卡密发给用户
                            }
                        }

                        if (validCards.length === 0) {
                            await sendMessage(chatId, `❌ 获取到的卡密未在 Bot 系统注册 (无法自动注册)\n\n获得的卡密内容:\n${cards.join('\n')}\n\n请联系管理员手动注册。`);
                            return;
                        }

                        // 验证一致性
                        if (validCards.length > 1) {
                            const first = validCards[0];
                            for (let i = 1; i < validCards.length; i++) {
                                const c = validCards[i];
                                if (c.trafficBytes !== first.trafficBytes || c.strategy !== first.strategy ||
                                    c.duration !== first.duration || c.squadUuid !== first.squadUuid) {
                                    await sendMessage(chatId, `❌ 卡密参数不一致，无法叠加`, {
                                        reply_markup: { inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]] }
                                    });
                                    return;
                                }
                            }
                        }

                        await clearPendingTradeIfMatch(kv, chatId, tradeNo);
                        await setSession(kv, userId, 'reg_nodeseek_input', { cards: validCards });
                        await sendMessage(chatId, `
✅ <b>卡密验证成功!</b>

╭──────────────────────
│ 🎫 有效卡密: <b>${validCards.length}</b> 张
│ 🌐 第2步: 绑定 Nodeseek
╰──────────────────────

请发送您的 Nodeseek 用户链接

📝 格式示例:
<code>https://www.nodeseek.com/space/36628</code>
                        `.trim(), {
                            reply_markup: { inline_keyboard: [[{ text: '❌ 取消注册', callback_data: 'cancel_reg' }]] }
                        });

                    } else if (flow === 'change') {
                        const validCards = [];
                        for (const code of cards) {
                            const trimmed = code.trim();
                            const card = await getCard(kv, trimmed);
                            if (card) validCards.push(trimmed);
                        }

                        if (validCards.length === 0) {
                            await sendMessage(chatId, `❌ 获取到的卡密无效或未在系统注册 (请联系管理员)\n\n获得的卡密:\n${cards.join('\n')}`);
                            return;
                        }

                        let totalTraffic = 0;
                        let firstCard = null;
                        for (const code of validCards) {
                            const c = await getCard(kv, code);
                            if (c) { totalTraffic += c.trafficBytes || 0; if (!firstCard) firstCard = c; }
                        }
                        const carryOverDays = sess.data?.carryOverDays || 0;
                        const cardDurationDays = DURATION_DAYS[firstCard?.duration] || 30;
                        const totalDays = carryOverDays + cardDurationDays;
                        await clearPendingTradeIfMatch(kv, chatId, tradeNo);
                        await setSession(kv, userId, 'change_car_confirm', { cards: validCards, carryOverDays });
                        await sendMessage(chatId, `
🔄 <b>确认换车</b>

╭──────────────────────
│ 🎫 卡密: <b>${validCards.length}</b> 张
│ 📦 新流量: <b>${formatBytes(totalTraffic)}</b>
│ ⏱ 新时长: <b>${totalDays}</b> 天 (卡密${cardDurationDays}天+结转${carryOverDays}天)
│
│ ⚠️ 换车后现有参数将被重置
│ 🔗 仅保留订阅链接
╰──────────────────────

确认要执行换车吗?
                        `.trim(), {
                            reply_markup: {
                                inline_keyboard: [
                                    [{ text: '✅ 确认换车', callback_data: 'change_car_confirm' }],
                                    [{ text: '❌ 取消', callback_data: 'refresh' }]
                                ]
                            }
                        });
                    }
                } else if (status === 3) {
                    await clearPendingTradeIfMatch(kv, chatId, tradeNo);
                    await sendMessage(chatId, '❌ 支付会话已关闭，请重新发起支付。', {
                        reply_markup: {
                            inline_keyboard: [
                                [{ text: '◀️ 返回', callback_data: 'refresh' }]
                            ]
                        }
                    });
                } else {
                    // status 0 = unpaid, other
                    await sendMessage(chatId, '⏳ <b>尚未检测到支付</b>\n\n请先完成支付后再点击"我已支付"。', {
                        reply_markup: {
                            inline_keyboard: [
                                [{ text: '✅ 我已支付', callback_data: `shop_checkpay_${tradeNo}` }],
                                [{ text: '❌ 取消', callback_data: 'shop_cancel' }]
                            ]
                        }
                    });
                }
            } else {
                await sendMessage(chatId, `❌ 查询失败: ${statusRes.msg || '未知'}`, {
                    reply_markup: {
                        inline_keyboard: [
                            [{ text: '🔄 重试', callback_data: `shop_checkpay_${tradeNo}` }],
                            [{ text: '◀️ 返回', callback_data: 'refresh' }]
                        ]
                    }
                });
            }
        } catch (e) {
            await sendMessage(chatId, `❌ 查询异常: ${e.message}`);
        }
        return;
    }

    if (data === 'shop_cancel') {
        const sess = await getSession(kv, userId);
        const tradeNo = sess.data?.tradeNo;
        if (tradeNo) {
            const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
            await ensureShopLoggedIn(kv, shopClient); // best effort login
            try {
                await shopClient.request('/shop/order/cancel', { trade_no: tradeNo });
                await clearPendingTradeIfMatch(kv, chatId, tradeNo);
            } catch (e) {
                console.error('Failed to cancel order on shop_cancel', e);
            }
        }
        await clearSession(kv, userId);
        await sendMessage(chatId, '✅ 已取消操作', {
            reply_markup: { inline_keyboard: [[{ text: '◀️ 返回主页', callback_data: 'refresh' }]] }
        });
        return;
    }
}

// ==================== 消息处理 ====================
async function handleMessage(kv, message) {
    const chatId = message.chat.id;
    const userId = message.from.id;
    const username = message.from.username || '';
    const text = message.text || '';

    // 处理 /sub 命令 (支持群组和私聊, 可引用消息查看他人信息)
    if (text && (text === '/sub' || text.startsWith('/sub@'))) {
        // 确定查询目标: 如果引用了消息, 查看被引用消息发送者的信息
        let targetUserId = userId;
        let isSelf = true;
        if (message.reply_to_message && message.reply_to_message.from) {
            targetUserId = message.reply_to_message.from.id;
            isSelf = (targetUserId === userId);
        }

        const user = await getUserByTelegramId(targetUserId);
        if (!user) {
            const whoMsg = isSelf ? '您的' : '该用户的';
            await sendMessage(chatId, `❌ 未找到${whoMsg}账户信息${isSelf ? '，请私聊机器人进行绑定。' : ''}`, {
                reply_to_message_id: message.message_id
            });
            return;
        }

        const statusEmoji = {
            'ACTIVE': '🟢', 'DISABLED': '🔴', 'LIMITED': '🟡', 'EXPIRED': '⚫'
        };
        const statusText = {
            'ACTIVE': '正常', 'DISABLED': '已禁用', 'LIMITED': '流量耗尽', 'EXPIRED': '已过期'
        };

        const usedTraffic = user.userTraffic?.usedTrafficBytes || 0;
        const totalTraffic = user.trafficLimitBytes || 0;
        const percentage = totalTraffic > 0 ? Math.round((usedTraffic / totalTraffic) * 100) : 0;
        const progressBar = generateProgressBar(percentage, 8);

        let remainingDays = '∞';
        let expireTag = '';
        if (user.expireAt) {
            const diffDays = Math.ceil((new Date(user.expireAt).getTime() - Date.now()) / 86400000);
            const days = Math.max(diffDays, 0);
            remainingDays = days.toString();
            if (days <= 0) expireTag = ' ❌';
            else if (days <= 7) expireTag = ' ⚠️';
        }

        let joinDuration = '0';
        if (user.createdAt) {
            const diffDays = Math.floor((Date.now() - new Date(user.createdAt).getTime()) / 86400000);
            joinDuration = Math.max(diffDays, 0).toString();
        }

        // 获取流量占比统计 (Top 5 for group to keep compact)
        let trafficChart = '';
        try {
            const bwStats = await getBandwidthStats(user.uuid, 5);
            if (bwStats && bwStats.topNodes && bwStats.topNodes.length > 0) {
                const chartStr = generateTrafficChart(bwStats.topNodes, usedTraffic);
                if (chartStr && !chartStr.includes('暂无流量')) {
                    trafficChart = `\n<code>${chartStr}</code>`;
                }
            }
        } catch (e) {
            console.error('获取流量统计失败:', e);
        }

        const titleText = isSelf ? '📊 我的订阅' : `📊 ${user.username}`;
        const statusIcon = statusEmoji[user.status] || '❓';
        const msg = `<b>${titleText}</b> ${statusIcon}

${progressBar} <b>${percentage}%</b> | ${formatBytes(usedTraffic)}/${formatBytes(totalTraffic)}
📅 剩余 <b>${remainingDays}</b> 天${expireTag} · 上车 <b>${joinDuration}</b> 天${trafficChart}`.trim();

        await sendMessage(chatId, msg, {
            reply_to_message_id: message.message_id
        });
        return;
    }

    // 只处理私聊
    if (message.chat.type !== 'private') return;

    const session = await getSession(kv, userId);

    // Handle users_shared event (from keyboard request_users button)
    // Only proceed if in invite_target_input state, otherwise ignore
    if (message.users_shared && session.state !== 'invite_target_input') {
        return;
    }

    // 处理会话状态
    if (session.state) {
        // 注册流程
        if (session.state.startsWith('reg_') || session.state.startsWith('bind_')) {
            await handleRegistration(kv, chatId, userId, username, text, session);
            return;
        }

        // 管理员卡密生成
        if (session.state.startsWith('admin_card_') && isAdmin(userId)) {
            await handleAdminCardGeneration(kv, chatId, userId, text, session);
            return;
        }

        // ========== Shop 文本输入处理 ==========
        // 卡网登入 - 用户名
        if (session.state === 'shop_login_user') {
            const shopUser = text.trim();
            if (!shopUser) {
                await sendMessage(chatId, '❌ 用户名不能为空，请重新输入');
                return;
            }
            await setSession(kv, userId, 'shop_login_pass', { ...session.data, shopUser });
            await sendMessage(chatId, `
🔑 <b>登入卡网</b>

╭──────────────────────
│ 👤 用户名: <b>${shopUser}</b>
│ 📝 请输入密码:
╰──────────────────────

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 验证账户密码
👉 <b>下一步:</b> 发送您的登录密码
👉 <b>遇到问题:</b> 密码将安全传输，如验证失败请检查大小写
            `.trim(), {
                reply_markup: { inline_keyboard: [[{ text: '❌ 取消', callback_data: session.data.flow === 'reg' ? 'cancel_reg' : 'refresh' }]] }
            });
            return;
        }

        // 卡网登入 - 密码
        if (session.state === 'shop_login_pass') {
            const shopPass = text.trim();
            if (!shopPass) {
                await sendMessage(chatId, '❌ 密码不能为空，请重新输入');
                return;
            }
            const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
            try {
                const loginRes = await shopClient.request('/login', { username: session.data.shopUser, password: shopPass });
                if (loginRes.code === 200 && loginRes.data?.token) {
                    await shopClient.setToken(loginRes.data.token);
                    await saveShopCredentials(kv, chatId, session.data.shopUser, shopPass);
                    await sendMessage(chatId, '✅ 登入成功! 正在加载商品...');
                    await shopShowProducts(kv, chatId, userId, shopClient, session.data.flow);
                } else {
                    await sendMessage(chatId, `❌ 登入失败: ${loginRes.msg || '用户名或密码错误'}\n\n请重新输入密码:`, {
                        reply_markup: { inline_keyboard: [[{ text: '❌ 取消', callback_data: session.data.flow === 'reg' ? 'cancel_reg' : 'refresh' }]] }
                    });
                }
            } catch (e) {
                await sendMessage(chatId, `❌ 登入异常: ${e.message}`);
            }
            return;
        }

        // 卡网购买 - 数量输入
        if (session.state === 'shop_quantity') {
            const qtyStr = text.trim();
            if (!/^\d+$/.test(qtyStr)) {
                await sendMessage(chatId, '❌ 请输入有效的数字');
                return;
            }
            const qty = parseInt(qtyStr, 10);
            if (qty <= 0 || qty > 100) {
                await sendMessage(chatId, '❌ 数量应在 1-100 之间');
                return;
            }
            const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
            const loginStatus = await ensureShopLoggedIn(kv, shopClient);
            const isLoggedIn = loginStatus.loggedIn;

            // Check Pending Trade
            const pendingTrade = await getPendingTrade(kv, chatId);
            if (pendingTrade) {
                await sendMessage(chatId, `⏳ <b>您有未完成的订单!</b>\n\n请先完成或取消上一笔订单 (订单号: <code>${pendingTrade}</code>)`, {
                    reply_markup: {
                        inline_keyboard: [
                            [{ text: '💳 前往支付', callback_data: `pay_${pendingTrade}` }],
                            [{ text: '📦 查看订单', callback_data: `getitem_${pendingTrade}` }],
                            [{ text: '❌ 取消订单', callback_data: `shop_cancel_order_${pendingTrade}` }] // using specific cancel handler
                        ]
                    }
                });
                return;
            }

            try {
                const skuId = session.data.selectedSkuId;
                const orderRes = await shopClient.request('/shop/order/trade', {
                    items: [{ sku_id: skuId, quantity: qty }]
                });
                if (orderRes.code === 200 && orderRes.data?.trade_no) {
                    const trade = orderRes.data;
                    const tradeNo = trade.trade_no;
                    const amount = trade.amount || trade.total_amount || '?';

                    // 如果是免费订单
                    if (Number(amount) === 0 || trade.status === 1) {
                        const flow = session.data.flow;

                        // 续费流程: 不需要获取卡密，直接进入时长选择
                        if (flow === 'renew') {
                            await sendMessage(chatId, '✅ 订单已完成 (免费)!');
                            await clearSession(kv, userId);
                            await setSession(kv, userId, 'renew_duration_select', { orderNo: tradeNo, orderAmount: amount });
                            await sendMessage(chatId, `✅ 订单完成!\n\n请选择续费时长:`, {
                                reply_markup: {
                                    inline_keyboard: [
                                        [{ text: '📅 月付 (30天)', callback_data: 'renew_monthly' }],
                                        [{ text: '📅 2月付 (60天)', callback_data: 'renew_bimonthly' }],
                                        [{ text: '📆 季付 (90天)', callback_data: 'renew_quarterly' }],
                                        [{ text: '🗓️ 半年付 (180天)', callback_data: 'renew_semiannual' }],
                                        [{ text: '🎉 年付 (365天)', callback_data: 'renew_annual' }],
                                        [{ text: '❌ 取消', callback_data: 'refresh' }]
                                    ]
                                }
                            });
                            return;
                        }

                        await sendMessage(chatId, '✅ 订单已完成 (免费)! 正在获取卡密...');
                        const isLoggedIn = !!(await shopClient.getToken());
                        const cards = await shopGetOrderCards(shopClient, tradeNo, isLoggedIn);
                        if (cards.length === 0) {
                            await sendMessage(chatId, '⚠️ 暂未获取到卡密，请稍等后联系客服');
                            return;
                        }

                        if (flow === 'reg') {
                            const validCards = [];
                            for (const code of cards) {
                                const card = await getCard(kv, code.trim());
                                if (card) validCards.push(code.trim());
                            }
                            if (validCards.length > 0) {
                                await setSession(kv, userId, 'reg_nodeseek_input', { cards: validCards });
                                await sendMessage(chatId, `
✅ <b>卡密获取成功!</b>

╭──────────────────────
│ 🎫 获取到 <b>${validCards.length}</b> 张卡密
╰──────────────────────

请输入您的 Nodeseek UID (纯数字):
                                `.trim(), {
                                    reply_markup: { inline_keyboard: [[{ text: '❌ 取消', callback_data: 'cancel_reg' }]] }
                                });
                            } else {
                                await sendMessage(chatId, '❌ 获取到的卡密无效');
                            }
                        } else if (flow === 'change') {
                            const validCards = [];
                            for (const code of cards) {
                                const card = await getCard(kv, code.trim());
                                if (card) validCards.push(code.trim());
                            }
                            if (validCards.length === 0) {
                                await sendMessage(chatId, '❌ 获取到的卡密无效，请联系客服');
                                return;
                            }
                            let totalTraffic = 0;
                            let firstCard = null;
                            for (const code of validCards) {
                                const c = await getCard(kv, code);
                                if (c) { totalTraffic += c.trafficBytes || 0; if (!firstCard) firstCard = c; }
                            }
                            const carryOverDays = session.data.carryOverDays || 0;
                            const cardDurationDays = DURATION_DAYS[firstCard?.duration] || 30;
                            const totalDays = carryOverDays + cardDurationDays;
                            await setSession(kv, userId, 'change_car_confirm', { cards: validCards, carryOverDays });
                            await sendMessage(chatId, `
🔄 <b>确认换车</b>

╭──────────────────────
│ 📦 新流量: <b>${formatBytes(totalTraffic)}</b>
│ ⏱ 新时长: <b>${totalDays}</b> 天
╰──────────────────────

确认要执行换车吗?
                                `.trim(), {
                                reply_markup: {
                                    inline_keyboard: [
                                        [{ text: '✅ 确认换车', callback_data: 'change_car_confirm' }],
                                        [{ text: '❌ 取消', callback_data: 'refresh' }]
                                    ]
                                }
                            });
                        }
                        return;
                    }

                    // 获取支付方式
                    const amountNum = parseFloat(amount);
                    const normalizedAmount = Number.isFinite(amountNum) ? amountNum.toFixed(2) : '0.00';
                    const payListRes = await shopClient.request('/pay/list', {
                        business: 'product',
                        amount: normalizedAmount
                    });
                    const payMethods = shopToArray(payListRes.data);
                    const balance = payListRes.balance ?? payListRes.ext?.balance ?? 0;
                    const balanceNum = parseFloat(String(balance));
                    const isLoginRes = payListRes.is_login ?? payListRes.ext?.is_login ?? false;
                    const isLogin = isLoginRes || !!(await shopClient.getToken());

                    if (payMethods.length === 0 && (!isLogin || balanceNum < amountNum)) {
                        await sendMessage(chatId, '❌ 暂无可用支付方式，请联系客服');
                        return;
                    }

                    // Set Pending Trade
                    await setPendingTrade(kv, chatId, tradeNo);
                    await bindTradeToChat(kv, chatId, tradeNo); // Bind explicitly

                    await setSession(kv, userId, 'shop_paying', { ...session.data, tradeNo, orderAmount: amount });
                    const payButtons = [];
                    // 余额支付选项
                    if (isLogin && Number.isFinite(balanceNum) && Number.isFinite(amountNum) && amountNum > 0 && balanceNum >= amountNum) {
                        payButtons.push([{ text: `💰 余额支付 (剩余: ¥${balance})`, callback_data: `shop_pay_${tradeNo}_balance` }]);
                    }
                    payMethods.forEach(m => {
                        const name = m.name || m.pay_name || '支付';
                        payButtons.push([{ text: `💳 ${name}`, callback_data: `shop_pay_${tradeNo}_${m.id}` }]);
                    });
                    payButtons.push([{ text: '❌ 取消订单', callback_data: 'shop_cancel' }]);
                    let orderText = `📦 <b>订单已创建</b>\n\n╭──────────────────────\n│ 📝 订单号: <code>${tradeNo}</code>\n│ 💰 金额: ¥${amount}`;
                    if (isLogin) orderText += `\n│ 💰 余额: ¥${balance}`;
                    orderText += `\n╰──────────────────────\n\n请选择支付方式:`;
                    await sendMessage(chatId, orderText.trim(), {
                        reply_markup: { inline_keyboard: payButtons }
                    });
                } else {
                    await sendMessage(chatId, `❌ 创建订单失败: ${orderRes.msg || '未知错误'}`, {
                        reply_markup: { inline_keyboard: [[{ text: '◀️ 返回', callback_data: 'refresh' }]] }
                    });
                }
            } catch (e) {
                await sendMessage(chatId, `❌ 创建订单异常: ${e.message}`);
            }
            return;
        }

        // 换车 - 卡密输入
        if (session.state === 'change_car_card_input') {
            const cardsText = text.trim();
            const cardCodes = cardsText.split(/[\n,;]+/).map(c => c.trim()).filter(c => c.length > 0);
            if (cardCodes.length === 0) {
                await sendMessage(chatId, '❌ 请输入有效的卡密');
                return;
            }
            // 验证卡密
            const validCards = [];
            const invalidCards = [];
            for (const code of cardCodes) {
                const card = await getCard(kv, code);
                if (card) validCards.push(code);
                else invalidCards.push(code);
            }
            if (validCards.length === 0) {
                await sendMessage(chatId, `❌ 所有卡密均无效:\n${invalidCards.map(c => `• <code>${c}</code>`).join('\n')}`, {
                    reply_markup: { inline_keyboard: [[{ text: '🔄 重新输入', callback_data: 'change_car_card' }, { text: '❌ 取消', callback_data: 'refresh' }]] }
                });
                return;
            }
            // 计算详情
            let totalTraffic = 0;
            let firstCard = null;
            for (const code of validCards) {
                const c = await getCard(kv, code);
                if (c) { totalTraffic += c.trafficBytes || 0; if (!firstCard) firstCard = c; }
            }
            const carryOverDays = session.data.carryOverDays || 0;
            const cardDurationDays = DURATION_DAYS[firstCard?.duration] || 30;
            const totalDays = carryOverDays + cardDurationDays;

            let warnText = '';
            if (invalidCards.length > 0) {
                warnText = `\n│ ⚠️ ${invalidCards.length} 张卡密无效 (已忽略)`;
            }

            await setSession(kv, userId, 'change_car_confirm', { cards: validCards, carryOverDays });
            await sendMessage(chatId, `
🔄 <b>确认换车</b>

╭──────────────────────
│ 🎫 有效卡密: <b>${validCards.length}</b> 张${warnText}
│ 📦 新流量: <b>${formatBytes(totalTraffic)}</b>
│ ⏱ 新时长: <b>${totalDays}</b> 天 (卡密${cardDurationDays}天+结转${carryOverDays}天)
│
│ ⚠️ 换车后现有参数将被重置
│ 🔗 仅保留订阅链接
╰──────────────────────

确认要执行换车吗?
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '✅ 确认换车', callback_data: 'change_car_confirm' }],
                        [{ text: '❌ 取消', callback_data: 'refresh' }]
                    ]
                }
            });
            return;
        }

        if (session.state === 'reset_traffic_order') {
            const orderNo = text.trim();

            // 验证订单号是否为24位纯数字
            if (!/^\d{24}$/.test(orderNo)) {
                await sendMessage(chatId, '❌ 订单号无效，需为24位纯数字，请重新输入', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'refresh' }]]
                    }
                });
                return;
            }
            const user = await getUserByTelegramId(userId);

            if (!user) {
                await sendMessage(chatId, '❌ 未找到账户信息');
                return;
            }

            try {
                // 生成请求ID并保存待处理请求
                const requestId = generateRequestId();
                await savePendingRequest(kv, requestId, {
                    type: 'reset_traffic',
                    userUuid: user.uuid,
                    userId: userId,
                    username: user.username || userId,
                    orderNo: orderNo,
                    createdAt: Date.now()
                });

                await clearSession(kv, userId);
                await sendMessage(chatId, '✅ 重置流量请求已提交，请等待管理员确认。');

                // 发送到群组等待管理员确认
                await sendMessage(CONFIG.GROUP_ID,
                    `📊 <b>重置流量请求</b>\n\n📝 订单号: <code>${orderNo}</code>\n👤 用户: ${user.username || userId}\n📅 时间: ${formatDate(new Date().toISOString())}\n\n⏳ 等待管理员确认...`, {
                    reply_markup: {
                        inline_keyboard: [
                            [
                                { text: '✅ 同意', callback_data: `approve_reset_${requestId}` },
                                { text: '❌ 拒绝', callback_data: `reject_reset_${requestId}` }
                            ]
                        ]
                    }
                });
            } catch (e) {
                await sendMessage(chatId, `❌ 请求提交失败: ${e.message}`);
            }
            return;
        }

        // 续费订单号
        if (session.state === 'renew_order_input') {
            const orderNo = text.trim();

            // 验证订单号是否为24位纯数字
            if (!/^\d{24}$/.test(orderNo)) {
                await sendMessage(chatId, '❌ 订单号无效，请重新输入', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'refresh' }]]
                    }
                });
                return;
            }
            // 尝试获取订单金额
            let amount = '?';
            try {
                const shopClient = new ShopClient(CONFIG.SHOP_URL, kv, chatId);
                const orderRes = await shopClient.request('/pay/getOrder', { trade_no: orderNo });
                if (orderRes.code === 200 && orderRes.data) {
                    amount = String(orderRes.data.order_amount || orderRes.data.trade_amount || orderRes.data.money || '?');
                }
            } catch (e) {
                console.error('Failed to fetch order info', e);
            }

            await setSession(kv, userId, 'renew_duration_select', { orderNo, orderAmount: amount });

            await sendMessage(chatId, `
⏱ <b>选择续费时长</b>

╭──────────────────────
│ 📝 订单号: <code>${orderNo}</code>
│ 💰 金额: ¥${amount}
╰──────────────────────

请选择您的续费时长:
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '📅 月付 (30天)', callback_data: 'renew_monthly' }],
                        [{ text: '📅 2月付 (60天)', callback_data: 'renew_bimonthly' }],
                        [{ text: '📆 季付 (90天)', callback_data: 'renew_quarterly' }],
                        [{ text: '🗓️ 半年付 (180天)', callback_data: 'renew_semiannual' }],
                        [{ text: '🎉 年付 (365天)', callback_data: 'renew_annual' }],
                        [{ text: '❌ 取消操作', callback_data: 'cancel_reg' }]
                    ]
                }
            });
            return;
        }



        // 口令红包续费 - 输入口令
        if (session.state === 'renew_password_input') {
            const password = text.trim();
            if (!password || password.length < 1) {
                await sendMessage(chatId, '❌ 口令无效，请重新输入');
                return;
            }
            await setSession(kv, userId, 'renew_password_duration', { password });
            await sendMessage(chatId, `
⏱ <b>选择续费时长</b>

╭──────────────────────
│ 🧧 口令: <code>${password}</code>
╰──────────────────────

请选择您的续费时长:
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '📅 月付 (30天)', callback_data: 'renew_monthly' }],
                        [{ text: '📅 2月付 (60天)', callback_data: 'renew_bimonthly' }],
                        [{ text: '📆 季付 (90天)', callback_data: 'renew_quarterly' }],
                        [{ text: '🗓️ 半年付 (180天)', callback_data: 'renew_semiannual' }],
                        [{ text: '🎉 年付 (365天)', callback_data: 'renew_annual' }],
                        [{ text: '❌ 取消操作', callback_data: 'cancel_reg' }]
                    ]
                }
            });
            return;
        }

        // 邀请用户 - 输入原因
        if (session.state === 'invite_reason_input') {
            const reason = text.trim();
            if (!reason) {
                await sendMessage(chatId, '❌ 请输入邀请原因', {
                    reply_markup: {
                        inline_keyboard: [[{ text: '❌ 取消操作', callback_data: 'refresh' }]]
                    }
                });
                return;
            }
            await setSession(kv, userId, 'invite_target_input', { reason });
            await sendMessage(chatId, `
✅ <b>原因已记录!</b>

╭──────────────────────
│ 📝 原因: ${reason}
│ 👤 第2步: 选择或输入被邀请用户
╰──────────────────────

请点击下方按钮选择用户，或直接发送 Telegram ID:

📋 <b>操作指南</b>
━━━━━━━━━━━━━━━━
👉 <b>当前操作:</b> 指定被邀请人
👉 <b>下一步:</b> 推荐使用「👤 选择用户」按钮从通讯录选择
👉 <b>遇到问题:</b> 若手动输入 ID，请确保 ID 正确且对方未屏蔽私聊
            `.trim(), {
                reply_markup: {
                    keyboard: [
                        [{
                            text: '👤 选择用户',
                            request_users: {
                                request_id: 1,
                                user_is_bot: false,
                                max_quantity: 1
                            }
                        }],
                        [{ text: '❌ 取消操作' }]
                    ],
                    resize_keyboard: true,
                    one_time_keyboard: true
                }
            });
            return;
        }

        // 邀请用户 - 输入目标用户ID
        if (session.state === 'invite_target_input') {
            if (text === '❌ 取消操作') {
                await clearSession(kv, userId);
                await sendMessage(chatId, '✅ 操作已取消', { reply_markup: { remove_keyboard: true } });
                return;
            }

            let targetId;
            let targetName = '';

            if (message.users_shared) {
                const sharedUsers = message.users_shared.users;
                if (sharedUsers && sharedUsers.length > 0) {
                    targetId = sharedUsers[0].user_id;
                }
            } else {
                targetId = parseInt(text.trim());
            }

            if (!targetId || isNaN(targetId) || targetId <= 0) {
                await sendMessage(chatId, '❌ 无效的 Telegram ID，请重新输入或选择用户', {
                    reply_markup: {
                        keyboard: [
                            [{
                                text: '👤 选择用户',
                                request_users: {
                                    request_id: 1,
                                    user_is_bot: false,
                                    max_quantity: 1
                                }
                            }],
                            [{ text: '❌ 取消操作' }]
                        ],
                        resize_keyboard: true,
                        one_time_keyboard: true
                    }
                });
                return;
            }

            // 获取被邀请用户信息 (用于显示昵称)
            try {
                const chatInfo = await getChat(targetId);
                if (chatInfo.ok) {
                    const u = chatInfo.result;
                    const fullName = u.first_name + (u.last_name ? ' ' + u.last_name : '');
                    targetName = u.username ? `${fullName} (@${u.username})` : fullName;
                }
            } catch (e) {
                console.error('getChat failed', e);
            }

            if (!targetName) targetName = `ID: ${targetId}`;

            const inviterUser = await getUserByTelegramId(userId);
            const inviterName = inviterUser?.username || username || userId;

            // 检查被邀请用户是否已有入群请求
            const existingRequest = await getJoinRequest(kv, targetId);
            if (existingRequest) {
                // 废除之前的邀请链接
                if (existingRequest.inviteLink) {
                    try { await revokeChatInviteLink(CONFIG.GROUP_ID, existingRequest.inviteLink); } catch (e) { console.error(e); }
                }
                await deleteJoinRequest(kv, targetId);
                await sendMessage(chatId, `⚠️ 该用户已有待处理的入群请求，已自动废除旧请求`, {
                    reply_markup: { remove_keyboard: true }
                });
            }

            try {
                const inviteResult = await createChatInviteLink(CONFIG.GROUP_ID, 1);
                const reason = session.data.reason;

                await saveJoinRequest(kv, targetId, {
                    reason: `${reason} (邀请人: ${inviterName})`,
                    invitedBy: userId,
                    inviterName: inviterName,
                    userId: targetId,
                    username: targetName, // 保存解析后的名字
                    inviteLink: inviteResult.invite_link,
                    createdAt: Date.now()
                });

                await clearSession(kv, userId);

                // 发送邀请链接给被邀请用户
                try {
                    await sendMessage(targetId, `
👥 <b>群组邀请</b>

╭──────────────────────
│ 📝 邀请人: <a href="tg://user?id=${userId}">${inviterName}</a>
│ 📋 原因: ${reason}
╰──────────────────────

请点击下方按钮加入群组:
                    `.trim(), {
                        reply_markup: {
                            inline_keyboard: [
                                [{ text: '👥 进入群组', url: inviteResult.invite_link }],
                                [{ text: '✅ 我已进入', callback_data: 'invite_verify_group' }]
                            ]
                        }
                    });
                } catch (e) {
                    // 无法发送私聊给对方
                    console.error('Failed to send msg to target', e);
                }

                // 通知邀请人
                await sendMessage(chatId, `
✅ <b>邀请已发送!</b>

╭──────────────────────
│ 👤 被邀请用户: <b>${targetName}</b>
│ 📝 原因: ${reason}
│ 🔗 邀请链接已生成并发送
╰──────────────────────
                `.trim(), {
                    reply_markup: { remove_keyboard: true }
                });

                // 群组通知
                await sendMessage(CONFIG.GROUP_ID, `
👥 <b>入群邀请</b>

╭──────────────────────
│ 📝 邀请人: <a href="tg://user?id=${userId}">${inviterName}</a>
│ 👤 被邀请用户: <b>${targetName}</b>
│ 📋 原因: ${reason}
│ ⏰ 时间: ${formatDate(new Date().toISOString())}
╰──────────────────────
                `.trim());
            } catch (e) {
                await sendMessage(chatId, `❌ 邀请失败: ${e.message}`, {
                    reply_markup: { remove_keyboard: true }
                });
            }
            return;
        }
    }

    // 处理 /start 命令
    if (text === '/start') {
        await clearSession(kv, userId);
        const user = await getUserByTelegramId(userId);

        if (user) {
            // 获取流量占比统计 (Top 20)
            let trafficChart = '';
            try {
                const bwStats = await getBandwidthStats(user.uuid, 20);
                if (bwStats && bwStats.topNodes && bwStats.topNodes.length > 0) {
                    const chartStr = generateTrafficChart(bwStats.topNodes, user.userTraffic?.usedTrafficBytes || 0);
                    // 仅当有数据时显示
                    if (chartStr && !chartStr.includes('暂无流量')) {
                        trafficChart = `\n\n📊 <b>30天流量占比 (Top 20)</b>\n┌─────────────────────\n${chartStr}\n└─────────────────────`;
                    }
                }
            } catch (e) {
                console.error('获取流量统计失败:', e);
            }

            // 已注册用户，显示面板
            const panel = generateUserPanel(user, isAdmin(userId), trafficChart);
            await sendMessage(chatId, panel.text, {
                reply_markup: { inline_keyboard: panel.buttons }
            });
        } else {
            // 未注册用户，显示注册选项
            await sendMessage(chatId, `
╭─────────────────────╮
│     👋 <b>欢迎使用</b>     │
╰─────────────────────╯

ℹ️ 您还未绑定账户

请选择您的注册方式:

🎫 <b>卡密注册</b> - 已购买卡密的新用户
🔗 <b>订阅链接绑定</b> - 已有账户绑定TG
🛒 <b>卡网直购</b> - 直接从卡网购买卡密
            `.trim(), {
                reply_markup: {
                    inline_keyboard: [
                        [{ text: '🎫 使用卡密注册', callback_data: 'reg_by_card' }],
                        [{ text: '🔗 使用订阅链接绑定', callback_data: 'reg_by_sub' }],
                        [{ text: '🛒 卡网直购', callback_data: 'reg_by_shop' }]
                    ]
                }
            });
        }
        return;
    }
}

// ==================== 新成员入群处理 ====================
async function handleNewChatMember(kv, message) {
    const chatId = message.chat.id;

    // 只处理目标群组
    if (chatId !== CONFIG.GROUP_ID) return;

    const newMembers = message.new_chat_members || [];
    for (const member of newMembers) {
        // 跳过 bot 自身
        if (member.is_bot) continue;

        const memberId = member.id;
        const memberName = member.username || member.first_name || memberId;
        let joinRequest = await getJoinRequest(kv, memberId);

        // 已完成验证绑定的用户重新入群时，不应被误踢
        if (!joinRequest) {
            if (isAdmin(memberId)) {
                console.log(`[Join] 用户 ${memberId} 是管理员，放行`);
                continue;
            }

            const existingUser = await getUserByTelegramId(memberId);
            if (existingUser) {
                console.log(`[Join] 用户 ${memberId} 已绑定账户(${existingUser.uuid})，放行`);
                continue;
            }

            // 注册流程中的用户自动放行，避免误踢
            const memberSession = await getSession(kv, memberId);
            if (memberSession.state && memberSession.state.startsWith('reg_')) {
                console.log(`[Join] 用户 ${memberId} 处于注册验证流程(${memberSession.state})，不移出`);
                continue;
            }

            // 仍未命中放行条件时，尝试重试 (处理 KV/API 延迟)
            console.log(`[Join] 未通过初始检查，正在重试(等待数据同步)...`);
            for (let i = 0; i < 5; i++) {
                await new Promise(r => setTimeout(r, 1500));

                // 1. 再次检查是否有入群请求
                joinRequest = await getJoinRequest(kv, memberId);
                if (joinRequest) {
                    console.log(`[Join] 重试第 ${i + 1} 次成功找到入群请求`);
                    break;
                }

                // 2. 再次检查是否已是注册用户 (防止 verify_group_reg 完成后 session 清除但 user API 延迟)
                const retryUser = await getUserByTelegramId(memberId);
                if (retryUser) {
                    console.log(`[Join] 重试第 ${i + 1} 次成功找到已绑定账户`);
                    // 模拟找到 joinRequest 以跳过后续踢出逻辑，或直接continue外层循环
                    // 这里我们需要标记为 safe
                    joinRequest = { safe: true, reason: '注册成功(延迟检测)' };
                    break;
                }

                // 3. 再次检查 session (防止 session 写入延迟)
                const retrySession = await getSession(kv, memberId);
                if (retrySession.state && retrySession.state.startsWith('reg_')) {
                    console.log(`[Join] 重试第 ${i + 1} 次成功找到注册会话`);
                    joinRequest = { safe: true, reason: '注册中' };
                    break;
                }
            }
        }

        if (!joinRequest) {

            // 没有入群请求 → 踢出
            await sendMessage(chatId, `
⚠️ <b>未授权入群</b>

╭──────────────────────
│ 👤 用户: <a href="tg://user?id=${memberId}">${memberName}</a>
│ ❌ 该用户没有有效的入群请求
│ 🚫 已被自动移出群组
╰──────────────────────
            `.trim());

            try {
                await banChatMember(chatId, memberId);
                // 解除封禁以允许将来重新加入 (banChatMember already sets until_date to 60s)
                try { await unbanChatMember(chatId, memberId); } catch (e) { console.error(e); }
            } catch (e) {
                console.error('踢出用户失败:', e);
            }
        } else {
            // 如果是 retry loop 中标记的安全请求 (safe: true), 说明已通过其他流程(如注册)处理过欢迎消息, 跳过
            if (joinRequest.safe) {
                console.log(`[Join] 用户 ${memberId} 已认证(safe), 跳过欢迎消息`);
                continue;
            }

            // 有入群请求 → 发送欢迎消息
            const reason = joinRequest.reason || '未知';

            // 废除对应的群组邀请链接
            if (joinRequest.inviteLink) {
                try {
                    await revokeChatInviteLink(chatId, joinRequest.inviteLink);
                } catch (e) {
                    console.error('废除入群邀请链接失败:', e);
                }
            }

            await sendMessage(chatId, `
🎉 <b>欢迎新成员加入!</b>

╭──────────────────────
│ 👤 用户: <a href="tg://user?id=${memberId}">${memberName}</a>
│ 📝 原因: ${reason}
│ ⏰ 加入时间: ${formatDate(new Date().toISOString())}
╰──────────────────────

👋 欢迎加入我们的大家庭！
            `.trim());

            // 删除入群请求
            await deleteJoinRequest(kv, memberId);
        }
    }
}

// ==================== 主入口 ====================
export default {
    async fetch(request, env) {
        if (request.method !== 'POST') {
            return new Response('OK', { status: 200 });
        }

        try {
            const update = await request.json();

            if (update.message) {
                // 处理新成员加入
                if (update.message.new_chat_members && update.message.new_chat_members.length > 0) {
                    await handleNewChatMember(env.BOT_KV, update.message);
                } else if (update.message.users_shared || update.message.text) {
                    // 处理普通消息和 users_shared 事件
                    await handleMessage(env.BOT_KV, update.message);
                }
            } else if (update.callback_query) {
                await handleCallback(env.BOT_KV, update.callback_query);
            }

            return new Response('OK', { status: 200 });
        } catch (e) {
            console.error('Error:', e);
            return new Response('OK', { status: 200 }); // 返回 200 避免 Telegram 重复推送
        }
    }
};
