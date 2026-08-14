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

## User combo purchase interface display adjustment

- Display fulled internal squad as Grey/Disabled with "Full"/"满员" on it
- Display Provider name and its icon from Remnwave API for each node at the right of the card of the node
- Ban user from pressing continue button on coupon interface when the choosen combo could not be used
- Move "Confirm Purchase" text to the middle of the button

## Adjust combo renewal

Replace "Renew current combo" by "Automatic renewal". Move the text to the middle and display the button as green when auto renewal is on, red when off.
Auto renewal should default be on.
Display renewal price on the pop-up window, the date of auto renewal would perform, and the date of next cycle.
Display a switch button of whether auto renewal should be turned on.
User could not turn on auto renewal when they have no enough balance for next renewal, have queued combo, or currently unable to renew.
Ban user from entering combo purchase interface if auto renew is turned on.

When it comes to the end of the end of this billing cycle:
Case 1: user have no enough balance for renewal: Disable the subscription as normal same as auto renewal turned on
Case 2: user have enough balance: Auto deduct balance and renew the combo

Recurring discount should still be applied even when the coupon is disabled, edited, or out of usage count.
