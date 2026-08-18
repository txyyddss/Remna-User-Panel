# Your Goal

I need you to COMPLETE TASKS BELOW for a carpool panel called "TX Carpool". Visually, it need to be as simple as possible and using premium-dark design. Refer to the requirements below and follow the skills. The frontend should be designed initially for mobile device, and have tablet device support.

When coding, split the codes into main entrance and modules to ensure its extendability and maintainability. Write module specification when finish. Organize everything into folders, do not contain less than 2 files per folder, and do not contain more than 200 lines per file. Write an README.md under every frontend and backend sub-folers.

DO NOT hardcode any language into frontend code. Add them to the language files.
DO NOT store unnecessary infos in database, reuse the datas if possible.

The upstream project will be remnawave and emby. Follow their api documents STRICTLY. All api documents are located at ./reference. Make the best of the projects. For api calls performed by the backend, it must go into the queue first.

NEVER RUN TESTS LOCALLY. Audit the codes that have git differences currently when finished.

YOU MUST FOLLOW THE SKILLS (if there are conflicts, follow by the order): "I have ADHD", "Minimalist skill", "taste skill", "Telegram mini app", "nuxt-ui", "vue-best-practices", "golang-pro"
Use the MCPS for reference: "Nuxt UI", "Vue Docs"

Always use native components first, and then nuxt ui or other related projects. DO NOT build the wheel.

You may want to use the following projects for frontend:

- Nuxt UI v4
- Nuxt Icon
- Nuxt Fonts (Remember to apply different fonts for chinses and english)
- Shadcn
- Zod (for input validation)
- AutoAnimate
- TanStack Table
- Feel free to use other projects.

---

## Make adjustments / Fix bugs

- Change monthly reset cadence from MONTH to MONTH_ROLLING
- Remove the Coupon from user add coupon interface if the coupon is invalid or out of usage
- Add back button for Review purchase page
- Display checkin text in the middle of button on "Daily check-in" card

## Improve payment channel layout on Add TXB interface

- Should be the same style of Choose a payment provider
- Change the amount input to "InputNumber - With currency format" of NUXT UI
- Add icon to each payment channel, icon of wechatpay, alipay, and USDT are at [Paymentmethodlogo](d:/Remna-User-Panel/reference/Paymentmethodlogo/) 
- Display the tick at the right
- Remove "Choose a payment channel" text
- Admin could set a minimum and maximum TXB input, remember to limit on frontend and check in backend

## Add connections page and drop connection button

Display a "connections" link with a "link" button on "Subscription link" card, similar to the revoke button. When user enters the page, display a loading animation with progress bar (progress-percent in the response, lookup for multiple times if isCompleted is not true) on it. In backend, get connections by using Request Connections for User​ and Get Connections for User by Job ID​. Display multiple cards by different nodes(display their name and national flag), and display ip, and lastSeen, with a dropconnection button (unlink icon). Display a confirmation popup and call Drop Connections for Users or IPs​ when user proceeds.

## Add paid traffic reset

Add a traffic reset button (only icon, not text) at the right of renewal display card on Your ride.
Popup a confirmation window showing the reset price. Call remnawave api to reset traffic immidiately after confirm and payment.
The reset price should be calculated based on the ORIGINAL price of the combo user currently have.
For daily reset, the reset price is (ORIGINAL price/30), weekly reset (ORIGINAL price/4), monthly reset (ORIGINAL price).

## Improve miniapp layout

### Fullscreen mode

Display Hi, {telegram_first_name} \n @{username} in the top-middle. Make sure to use the space of reserve space.

### Not in fullscreen mode

Remove the reserve space on the top and do not display username.

## Hide Rollover forecast when user already exceeds rollover threshold

## Database

Do auto cleanup once per day. Backup before the job,

- Remove the session from database when it ends or expires
- Remove purchase, payment and daily checkin history which exists more than 7 days
- Remove group messages windows and events record which exists more than 24 hours

Also allow admin to upload database backup. Ensure it's validity before restoring the backup.

## Improve user editing admin interface

Admin would be able to perform following actions directly:

- View, edit, and refund their entitlements
- Perform bulk day extension by filtering combos or addons.
- View their emby accounts (combine emby account page into user page)
- View and edit their active combo

## Add Refund button

If user's current traffic = 0, time from purchase < 24h and not having renewal history, replace auto renewal button by refund and turn the button colour to yellow. Display the refund value in the popup. The refund price calculation should be the same as queued combo refund.

## Improve user combo storaging

Combine active purchases and purchase_addons into user profile.

## Add bot commands (all could perform in group)

DO NOT COUNT IN daily message reward when sending commands.
Automatically register the commands using telegram api.
User could perform by themselves or refer to a message sent by others to show.
- /sub - show their usage, e.g.:
⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️ 3% | 18.69 GB/550 GB
📅 剩余 423 天 · 上车 216 天
[▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░█] 84.2 GB

▓ US-AK (US) - 73.62 GB (87.4%)
░ LY-UK (GB) - 9.7 GB (11.5%)
█ LY-JP (JP) - 478.68 MB (0.6%)
▒ HN-TW (TW) - 268.31 MB (0.3%)
▇ AK-HK (HK) - 151.46 MB (0.2%)
- /balance - show the balance
- /signin - perform signin
- /start - Tell user to press "打开应用" or "Open TX Carpool" to use the app instead of using /start
- /mycombo - Show combo name, squads, rollover status (will get, will not get, cant get), traffic limit, reset cadence

## Update multiplier in node names

Per 30 minutes (perform with refresh statistics data), get all hosts and replace term like {0~9}.{0~9} in each host's name to the multiplier of linked nodes. If there are more than one node linked, skip it.

## Add Statistics page to "Around TX"

Refresh and cache the CALCULATED datas per 30 minutes (10 seconds for Get Nodes Metrics and Get nodes).
Use shadcn charts (use shadcn mcp for inquiries). The diaplay requirements are already given to you.
Cut long data names when displaying.

### Remnawave (RW)

1. Weekly user increase (Get Stats Digest​)
2. Online users of each node (Get Nodes Metrics​) & Node name & online status & rxBytesPerSec & txBytesPerSec & Xray version & multiplier (Get nodes​)
3. Monthly average usage percentage (same as pre-existed rollover calculation)
4. Per-day traffic usage of last seven days (Get Nodes Usage by Range, show each nodes in different colour), present as Mon,Tue,etc.

### Database (DB)

Because some of the records would be deleted in 7 days/24 hours,etc. due to the cleanup,you may not set the data date range intentionally.

1. Percentage of new user purchasing a combo
2. Average user spend
3. User spend range
4. Proportion of (Neither having auto renewal nor queued combo but have active combo) : (having auto renewal on) : (have a queued combo) : (user do not have an active combo)
5. Average rollover price
6. Daily checkin reward average
7. Percentage of user choosing each combo (count users who have a active combo)
8. Group messages total
9. Average number of Optional squads user would like to choose (count users who have a active combo)
10. Proportion of payments (Expired) : (Cancelled) : (Paid)
11. Current database size
12. Each percentage of each internal squads consists of each core combo (count users who have a active combo)

## Present using

- RW1+DB1: Pie Chart - Donut with Text
- RW2: Combine formatted datas of each node into cards
- RW4: Tooltip - Advanced
- RW3+DB2+DB3+DB5+DB6+DB8+DB9+DB12: Combine into a card using formatted datas
- DB4: Pie Chart - Label
- DB7: Pie Chart - Label List
- DB10: Pie Chart - Stacked (inner: BEPUSDT,outer: EZPay)
- DB12: Bar Chart - Stacked + Legend

## Improve compo purchase interface

- Display percentage of current user/Stock limit on Optional squads
- Display rollover threshold on Core combos

## Check auto renewal implementation
Check the program will renew/activate queued combo the subscription WITHOUT the interrupts of expiration on Remnawave. Perform changes 3 minutes before the remnawave marks it as expiration.

