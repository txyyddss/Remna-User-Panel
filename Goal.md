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
- Nuxt Charts (for displaying statistics)
- Nuxt Icon
- Nuxt Fonts (Remember to apply different fonts for chinses and english)
- Shadcn Nuxt
- Zod (for input validation)
- AutoAnimate
- TanStack Table (for database editing interface)
- Feel free to use other projects.

---

## Fix bugs

- Statistics displays wrongly on mobile
- Blance display of user home page is not at the middle of the card, it appears lower than it should be
- Bet wins appears same vibration as lose, they should be different

## Adjustment / New Features
- Could add several groups of EZPay and BEPusdt config with CUSTOMIZABLE channel name, the editor should be fully visualized
- Add a payment provider: coupon, used to exchange coupons related to balance but not discount
- Remove "Local payment rail" description of EZPay
- Add renew button for "Your ride" with a confirmation pop-up and user-adjustable number of periods to buy
- Show traffic should be displayed using graph, not just data
- Remember the progress and choices of user on "Explore" page, resume when the user switch back from other pages
- Remove refresh button on activity page
- Do not reset user input of stake amount when after a bet confirmation popup
- Apply different rollover algorithms for different Reset cadence and Validity days cases
- Block user from continuing when theres no Accessible nodes for current choices
- Add stock limit for Optional squads, count active instead of counting purchase numbers
- Add reset strategy display for Core combos section for users
- Add coupon content display for Coupon section for users
- Add more details for review page like rollover strategy, etc.
- Combine payments page into review
- Relax username format request to -36 characters long, ^[a-zA-Z0-9_-]+$

## Check
- Check if the program will deactivate user's subscription when the subscription expires