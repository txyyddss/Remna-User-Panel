# Your Goal

I need you to COMPLETE SEVERAL TASKS for a carpool panel called "TX Carpool". Visually, it need to be as simple as possible and using premium-dark design. Refer to the requirements below and follow the skills. The frontend should be designed initially for mobile device, and have tablet device support.

When coding, split the codes into main entrance and modules to ensure its extendability and maintainability. Write module specification when finish. Organize everything into folders, do not contain less than 2 files per folder, and do not contain more than 200 lines per file. Write an README.md under every frontend and backend sub-folers.

DO NOT hardcode any language into frontend code. Add them to the language files.
DO NOT store unnecessary infos in database, reuse the datas if possible.

The upstream project will be remnawave and emby. Follow their api documents STRICTLY. All api documents are located at ./reference. Make the best of the projects. For api calls performed by the backend, it must go into the queue first.

NEVER RUN TESTS LOCALLY. Audit the codes that have git differences when finished.

YOU MUST FOLLOW THE SKILLS (if there are conflicts, follow by the order): "I have ADHD", "Minimalist skill", "taste skill", "Telegram mini app", "nuxt-ui", "vue-best-practices", "golang-pro"
Use the MCPS for reference: "Nuxt UI", "Vue Docs"

Always use native components first, and then nuxt ui or other related projects. DO NOT build the wheel.

You may want to use the following projects for frontend:

- Nuxt UI v4
- Nuxt Charts (for displaying statistics)
- Nuxt Icon
- Nuxt Fonts (Remember to apply different fonts for chinses and english)
- Shadcn Nuxt
- Zod (for input validation)
- AutoAnimate
- TanStack Table (for database editing interface)
- Feel free to use other projects.

---

## Compiling issue

Fix the following compiling (quality check) issues

- npm run audit:structure did not pass due to lack of some README.md
- go test -race ./... did not pass due to timeout
- Domain coverage did not pass because the coverage is below 80%

## Payment landing page

When the payment fininishes, user will be redirected to the "redirecting url" provided by the program. Add a landing page displaying words like "payment confirmed, safe to go back to miniapp" with an green tick icon on it.

## UI Adjustment

### Fullscreen mode adjustment

Adjust the UI for full-screen mode of telegram miniapp. Do not contain anything in the top-right or top-left corner as it will be covered by telegram buttons.
The language switching button should be at the bottom with only a "translation" icon, show a language switching pop-up when user clicked it.
No need to show the Logo and name of TX Carpool.

### Home page redesign

The components should be in this arrangements, from top to bottom, only display the listed things:

- Available balance with a REAL add balance button (not a link to the balance page)
- Minimal subscription link display (no need to be covered) with a copy button
- Traffic statistics card (only display a progress bar and related number with different colour for different usage percentage), display a "?" button at the right (display a pop-up windows showing DETAILED per node usage of adjustable starting date and ending date)
- Your ride (active combo display with traffic reset strategy and squads)
- Around TX (questionnair & emby)

### Explore page redisign

Split the purchase flow into several processes

1. Select Core combos
2. Select Optional squads
3. Show accessible nodes (Names, National flags, traffic multiplier)
4. Select/Add Coupon
5. Review selection
6. Payment

Remove the description text on the top "One account-wide traffic budget. Optional squads follow the same term."

### Remove "Balance" page

Remove balance page and replace it with "Activity".
Coupon (wallet) should be moved to purchase flow.
Add balance/balance dispay should be moved to home page.
Recent activity should be removed.

### Bug fixes/Adjustments

- Description styles are not displayed on the user interface
- Add delete button for coupons
- Add reissue button for expired/failed payments
- Move messages per reward and Calander timezone to settings
- Default language should be telegram language
- Traffic limit input should automatically convert input with traffic units into bytes(like "10GB", "100MB", etc.)


### Visualization editor for Onboarding content

It should be total visualized instead of editing its json.
Use arrangable and editable cards with visable icon selection and colour selection interface.
Title of the agreement page should also be editable.

### Animations

Add more animations for a smoother user experience.