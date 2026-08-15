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

## Add Cache

- During startup, cache frontend assests in memory to reduce disk read (you must ensure it's SAFE and WILL NOT CAUSE MEMORY LEAKES)
- Cache and refresh the multiplier of each node PER 5 minutes

## Rewrite rollover

1. Remove all rollover TXB limit
2. The traffic of rollover should be cauculated based on the node usage data (same as the already implemented traffic usage bar), with the multiplier of each node. The time of data should be limited to the date of the start of current entitlement and the current date.
When user opened the rollover detail face, display the following information:

- Predict whether or not user will be able to get a rollover (current data used/already gone days*(already gone days+days left)), if no, suggest user to use how much traffic fewer in the next few days in order to get a rollover, if yes, predict how much they will get for rollover
- The maximum traffic user could use in order to get a rollover
- Already used traffic

Only calculate rollover ONLY when auto renewal is on. Display only the warning text on the rollover detail card if the user haven't enabled auto renewal. Give the balance to user BEFORE auto renewal so the system could use the balance to renew.

When finished, remember to edit related test cases.
 
## Add "hidden squad" feature

Admin could choose to enable and input a activation code when editing internal squad. No need to hide "hidden squad". When the user chose an hidden squad, display a popup window requiring user to input the activation code when he pressed "continue" button. If the user chose more than one hidden squad, require them to input activation code for each hidden squad one by one.

## Frontend element display adjustment

- Enlarge the success icon on balance coupon redeem success interface
- Move the warning icon of revoke subscription popup to the title in order to display less empty space
- Improve the group message reward display