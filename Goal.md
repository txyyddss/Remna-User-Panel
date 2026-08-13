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
- Shadcn
- Zod (for input validation)
- AutoAnimate
- TanStack Table (for database editing interface)
- Feel free to use other projects.

---

Improve internal squad description editing and display.
Admin should set 3 types of the squad: Broadband, China Optimized, or International Network (default)
For broadbrand, fill in the ISP, port (Mbps), dynamic or static (edit using switch button), and detailed location (like Miaoli, Taiwan).
For China Optimized, fill in the route of the three major carriors of china (CT,CU,CM), port (Mbps, could be unlimited), location (country only).
For International Network, fill in the port speed (Mbps, could be unlimited), location (country only), and upstream carriers (could be multiple, split using commas).
All of the types may also have extra descriptions that supports markdown.
Auto-generate the description displayed to user by the infos by admin.
Display the infos using icons and words as few as possible.
Use NUXT UI for the components.