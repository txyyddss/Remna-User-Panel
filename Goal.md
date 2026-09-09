# Your Goal

I need you to COMPLETE TASK(S) BELOW for a carpool panel called "TX Carpool". Arrange the order of the tasks and complete them ONE BY ONE. Visually, it need to be as simple as possible and using premium-dark design. Refer to the requirements below and follow the skills. The frontend should be designed initially for mobile device, and have desktop support.

When coding, split the codes into main entrance and modules to ensure its extendability and maintainability. Write module specification when finish. Organize everything into folders, do not contain less than 2 files per folder, and do not contain more than 200 lines per file. Write an README.md under every frontend and backend sub-folers.

DO NOT hardcode any text into frontend code. Add them to the language files.
DO NOT store unnecessary infos in database, reuse the datas if possible.

The upstream project will be remnawave and emby. Follow their api documents STRICTLY. All api documents are located at ./reference. Make the best of the projects. For api calls performed by the backend, it must go into the queue first.

NEVER RUN TESTS LOCALLY. Audit the codes you have edited when finished. Also check if you have completed ALL the tasks.

YOU MUST FOLLOW THE SKILLS (if there are conflicts, follow by the order): "I have ADHD", "Minimalist skill", "taste skill", "Telegram mini app", "nuxt-ui", "vue-best-practices", "golang-pro"
Use the MCPS for reference: "Nuxt UI", "Vue Docs"

Always use native components first, and then nuxt ui or other related projects. DO NOT build the wheel.

You may want to use the following projects for frontend:

- Nuxt UI v4
- Nuxt Icon
- Zod
- AutoAnimate
- TanStack Table

Test all frontend changes with Chrome Devtools MCP and constructed data when finish.
Push and commit to main branch when finish. Monitor actions run and fix issues when failed.

---

## Task1: Add page switching sliding animation

Display sliding animation when switching pages

- Previous page: Slide out
- Target page: Slide in
- Display both aimation synchronously with no overlapping or interrupting
- The direction of animation should follow and be the opposite of the one on navigation bar
- e.g. on phones, "explore" sliding in from right to left when switching from home, vice versa
- e.g. on desktop, "explore" sliding in from bottom to top when switching from home, vice versa
- Sliding up/down on desktop and left/right on phones

## Task2: Replace current Vue component to NUXT component

### Navigation bar on phones

Refer to NUXT UI docs "Tabs" - "With bottom tab bar"

### Markdown Editor on admin interface

Refer to NUXT UI docs "Editor"

## Task3: Improve sidebar display on desktop

- No longer display a "member" icon, display the real user avatar from telegram instead, remember to use the NUXT "avatar" component
- Do not display "Telegram member", display the telegram username instead, and cut long names
- Do not display the greeting on the top for desktop as it duplicates with the one on sidebar
- Default resize the sidebar into minimal width of displaying full wordings of sidebar items.

## Task4: Add the ability to disable geockeck for internal squad

Admin could disable geockeck (default on) on per-squad settings.
Display "Geockeck disabled." on geocheck popup on user interface.
Do not change other displays related to geocheck.

## Task5: Fix issues

- Add squad interface always appears empty on desktop
- It appears error when the new backend responce does not match the cached one on frontend