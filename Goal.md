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

## Task1: Fix bugs or make adjustments as follows

- telegram_payment_success_announcement is marked as failed even it is a successful call
- Delete admin payments page and move "Grant courtesy credit" and "Refund" into each user profile's payment record
- The greeting at the top (@{username}) should display user's inputed usernaame not their telegram username
- Combine multiple save buttons into one on "System settings" page
- The payment channel selection collides with amount input on add txb interface
- The connections page cannot work properly
- Multiplier did not display properly
- Remove "Use traffic multiplier" button - always don't use multiplier
- Automatic renewal and rollover detail failed to display
- The icon of traffic reset and refresh (TX statistics page) does not display at the middle of button
- The program did not edit multiplier display in hosts name properly
- Squad composition should display squad name instead of squad uuid
- Payment states card did not display properly on TX statistics page
- Live node metrics could not be loaded even when remnawave is online
- Monthly average usage and Users added this week always displays 0
- Replace "New user conversion rate" in the middle of the chart by "Users added this week" and delete the standalone "Users added this week" display
- Improve telegram notification display using telegram markdown v2

## Task2: Conduct a systematic, multi-pass code audit of the provided codebase with a focus on security, architecture, performance, and platform-native standards

Audit Categories & Priorities:

1. Critical Logic & Security (Highest Priority)

   - Vulnerabilities: OWASP Top 10, insecure state management, authentication/authorization leaks.
   - API Integrity: Frontend/Backend mismatches, payload validation gaps, missing error handling.
   - Edge Cases: Unhandled promises, race conditions, and state mutations.
   - Error handling: Error page, fallbacks, and security.
   - Business Logic Vulnerabilities: Any function that can make an abuse by user.
   - API Misuse: Any incorrect usage of upstream api.

2. Framework & SDK Optimization

   - Telegram SDK: Identify missing native features.
   - Nuxt UI: Replace redundant custom CSS/components with built-in Nuxt UI (or shadcn, or related projects) primitives, icons, and composables.
   - Domain Coverage: Flag missing domain models, logic gaps, or orphaned endpoints.

3. Code Hygiene & Architecture

   - Redundancy: Identify dead, orphaned, or debug code to safely delete.
   - DRY Principles: Locate repeated logic across components/composables and propose shared modules.
   - Mobile UX: Spot layout breaks, safe-area inset issues, touch-target sizing, and viewport overflow on mobile browsers.
   - Performance: Detect memory leaks, unnecessary reactivity overhead, large bundle imports, and slow database/API calls.