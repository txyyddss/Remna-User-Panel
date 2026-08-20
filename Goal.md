# Your Goal

I need you to COMPLETE TASKS BELOW for a carpool panel called "TX Carpool". Visually, it need to be as simple as possible and using premium-dark design. Refer to the requirements below and follow the skills. The frontend should be designed initially for mobile device, and have tablet device support.

When coding, split the codes into main entrance and modules to ensure its extendability and maintainability. Write module specification when finish. Organize everything into folders, do not contain less than 2 files per folder, and do not contain more than 200 lines per file. Write an README.md under every frontend and backend sub-folers.

DO NOT hardcode any language into frontend code. Add them to the language files.
DO NOT store unnecessary infos in database, reuse the datas if possible.

The upstream project will be remnawave and emby. Follow their api documents STRICTLY. All api documents are located at ./reference. Make the best of the projects. For api calls performed by the backend, it must go into the queue first.

NEVER RUN TESTS LOCALLY. Audit the codes you have edited when finished.

YOU MUST FOLLOW THE SKILLS (if there are conflicts, follow by the order): "I have ADHD", "Minimalist skill", "taste skill", "Telegram mini app", "nuxt-ui", "vue-best-practices", "golang-pro"
Use the MCPS for reference: "Nuxt UI", "Vue Docs", "shadcn"

Always use native components first, and then nuxt ui or other related projects. DO NOT build the wheel.

You may want to use the following projects for frontend:

- Nuxt UI v4
- Nuxt Icon
- Shadcn
- Zod
- AutoAnimate
- TanStack Table

---

## Task: Add Affiliate Function

Add a "Affiliate Centre" page that could enter from Around TX.

### Aff Rule

The aff reward should be split into several tiers.
Admin could set tier names, number of successful AFFs to upgrade to each tier, payback percentage, and the upgrade reward (could be COUPON, TXB reward, or subscription extention).
All of the aff reward value should be adjustable by admin, and could be switched on and off.
ONLY person who made a "Add TXB" payment should be count as a successful aff.

#### Commission

The user will receive percentage rewards from FIRST EVER Add TXB payment of the user he invited.
For different tiers, use different percentage.

### Links

Display a link like `https://t.me/example?start=12345678`, which "ttps://t.me/example" is the bot link, and 12345678 be the user telegram id. Write the inviter id into user database, and leave blank if no aff is made. There should not be overwrite in inviter id unless the id is invalid. Display a welcoming message with the inviter's username on it when user uses the invite link. Write into database before the open of miniapp when received `/start <userid>`. Calculate payback using settlemet-time tier. Prevent existing user to use the invite link.

### Database storaging

Add new_user (true/false) and inviter_id column for user table. Mark all existing user as false and leave inviter_id as blank during migration.

### Display

The "Affiliate Centre" page should display several parts.

### Link & count

Display a AFF link with a copy button, and display the total payback number, user registered, and the convert percentage (Success affs / user registered * 100%) below.

#### Tier & Tier progress

Display a progress bar concerning the progress of upgrading to next tier. In below, display current payback percentage and number of affs left in order to upgrade in the left, and display next tier's payback percentage and upgrade award (if it is coupon, display the coupon name, if it is TXB award or subscription extention, display the number). For user who already in the top tier, display a congratuation text in the middle instead.

### AFF Detail

Each row should display telegram username, registration date, payback time (if present), payback amount (if present).
Display default 5 recent users each page, and a page switch button.

### Notification

Send notification to the afflink owner for a successful AFF, jincluding the telegram username, time, and payback amount.
Also sent a congratuation message during tier upgrade.
Use emoji and Markdownv2 to present.