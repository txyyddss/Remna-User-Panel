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

## Fix bugs / Make adjustments

- Show the amount of TXB gain for coupon redeem on add TXB interface similar to other payment methods
- Move the page title of agreement to the right of warning icon on the top of the page
- Renew ride price calculation should include the recurring coupon used for purchase the combo
- DELETE acessible nodes editing as the porogram already gets accessible nodes directly from remnawave
- The recurring coupon used for renew should not count in the use limit
- One-time discount should not be calculated in combo renewal
- Add a telegram native back button to add TXB interface as the close button may be covered because of too many channels/providers
- Replace the black and small tick on Order cancelled interface of Add TXB to a big red error icon
- The mask of progress bar of Choose your combo interface should be green but not gray
- Show the reward in the popup of lucky draws