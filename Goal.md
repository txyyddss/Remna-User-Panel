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

## I need you to fix bugs or make adjustments

- On Add TXB interface, Coupon code input box conflicts with Redeem button
- The per-node traffic view should be displayed in chart rather than datas
- The terms number should be displayed as slider on "Renew your ride"
- The total TXB diaplay seems weird with the date display on "Renew your ride"
- The continue button on "Choose your combo." page should not be flowting
- The coupon of 90% recurring discount displayed as "Adds 90% of the balance" on "Coupon" page of "Choose your combo."
- When user come back from other page to "Review" of "Choose your combo.", it always displays "Calculating quote" and connot confirm purchase
- "Term" of "Review" of "Choose your combo." contains translation errors
- Catalog, internal squad, and bet statistics displays error on mobile devices
- No stock limit settings on Internal squads editing, but it is implemented on backend, add this
- Add dropdown menu for Eligible combo and squad ids of coupon editing selecting the items instead of filling in the id
- Cannot deduct or add TXB for user on admin interface
- Payment profiles SHOULD NOT be per-channel driven. It should be multi (or one) set(s) of EZPay and BEPUSDT Provider with customizable provider names. Admin can edit and choose to enable channels for the providers seperately. No duplicate configs should be made.
- The "Backups" of admin interface displays weird on mobile devices
- Add search for table names on "Database editor" interface