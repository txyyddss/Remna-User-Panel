# Your Goal

I need you to PERFORM AN UI REWRITE for a carpool panel called "TX Carpool". Visually, it need to be as simple as possible and using premium-dark design. Refer to the requirements below and follow the skill "I have ADHD", "taste skill", "minimalist design skill", "nuxt-ui", "golang-pro" and "vue-best-practices". The frontend should be designed initially for mobile device, and have tablet device support.

When coding, split the codes into main entrance and modules to ensure its extendability and maintainability. Write module specification when finish. Organize everything into folders, do not contain less than 2 files per folder, and do not contain more than 200 lines per file. Write an README.md under every frontend and backend sub-folers.

DO NOT hardcode any language into frontend code. Add them to the language files.
DO NOT store unnecessary infos in database, reuse the datas if possible.

The upstream project will be remnawave and emby. Follow their api documents strictly. All api documents are located at ./reference. Make the best of the projects. For api calls performed by the backend, it must go into the queue first.

No need to test locally but audit all the codes when finished. 

---

## New features

- Add delete button for Questionnairs (also delete related imports and participant records)
- Add delete button for betting games and lucky draw (also delete related record in database)
- Add delete button to each Synchronization job and Backups
- Add searching and filtering in Database editor
- Add detailed graph statistics for each betting and lucky draw
- Show usage count of the coupon on admin interface
- Support colour and font size for markdown support in combo and internal squad description editing
- Show statistics detailed statictics of combo and internal squads purchace
- Admin could link internal squads with one or more nodes of remnawave and display its user traffic multiplier and national flag



## Adujustments

- Default enable all emby folders, store DISABLED folders in database
- Optimize admin UI: place the cataogs into drop down menu instead of tile
- Delete "Users" and "Emby accounts"admin interface as it could be edited diretly using database editor
- Delete reward value editing in Activity page as it is a duplicate of the one in Settings
- Daily reward could be set to a range of random value
- When the user already have active subsctiption, warn the user the date of the new combo to take effect while purchasing
- Use an pop-up window instead of simple message to show the result of betting or lucky draw for users
- Move "Around TX" to the top, right behind the balance display
- Admin could edit the words on welcoming screen and automatically calculate the display time based on the word count
- Admin could set multiple agreements and its icon, user must agree more.
- Optimize the github actions file to maximize compiling speed

## Bug fixs

- Could not create lucky draw
- Emby Libraries settings did not sync to the emby server
- Betting always displays its default icon instead of custom one

## Database

- No need to store squads that have not been edited yet
- Delets webhook events record after is was used (like add balance)
- Only need to store combo_id but other info of the combo for purchases
- Only store id but not aggregate_id for outbox_jobs
- No need to record/store join invites
- Combine "combo_squads" to combos
- Cache database into memory, but you MUST ensure data safety.

