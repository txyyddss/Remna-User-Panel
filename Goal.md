# Your Goal

I need you to COMPLETE SEVERAL TASKS for a carpool panel called "TX Carpool". Visually, it need to be as simple as possible and using premium-dark design. Refer to the requirements below and follow the skills. The frontend should be designed initially for mobile device, and have tablet device support.

When coding, split the codes into main entrance and modules to ensure its extendability and maintainability. Write module specification when finish. Organize everything into folders, do not contain less than 2 files per folder, and do not contain more than 200 lines per file. Write an README.md under every frontend and backend sub-folers.

DO NOT hardcode any language into frontend code. Add them to the language files.
DO NOT store unnecessary infos in database, reuse the datas if possible.

The upstream project will be remnawave and emby. Follow their api documents STRICTLY. All api documents are located at ./reference. Make the best of the projects. For api calls performed by the backend, it must go into the queue first.

No need to test locally but audit all the codes when finished. Check if there are api mismatch between frontend and backend.

YOU MUST FOLLOW THE SKILLS (if there are conflicts, follow by the order): "I have ADHD", "Minimalist skill", "taste skill", "Telegram mini app", "nuxt-ui", "vue-best-practices", "golang-pro"
ALL THE SKILLS ARE AVAILABLE NOW.

---

## Frontend rewrite

Rewrite the frontend using NuxtUI V4 and the latest version of Vue and vite. If an an component exist in nuxt ui, DO NOT rewrite the wheel. When finish, rewrite the github actions workflow.

DO NOT hardcode any text in the code. ALWAYS use an placeholder and put the text in the language files.

ALL icons should NOT be generated locally. DELETE them if they already exist in the project files.

You may want to use the following projects:
- Nuxt Charts (for displaying statistics)
- Nuxt Icon
- Nuxt Fonts (Remember to apply different fonts for chinses and english)
- Shadcn Nuxt
- Zod (for input validation)
- AutoAnimate
- TanStack Table (for database editing interface)

Your design should be visually fluent with animations and advanced visual effects. Wordings shown on the UI should be short and easy-to-read.

## Backend security update

- Generate and check signitures for EVERY backend calls.
- Use regex to verify the validity of every input by user in backend.
- Every api calls of upstream projects should go to the queue first.

## Backend files spilting

If an file of backend contains more than 300 lines, split them into several files by its usage if possible.

## README
Add an README.md under every frontend and backend sub-folder describing what are the responsibilities of each file under that folder are. Update exiting docs and add links to the newly created READMEs as an index
