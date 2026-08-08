# Your Goal

I need you to add content for a carpool panel called "TX Carpool". Visually, it need to be as simple as possible and using premium-dark design. Refer to the requirements below and follow the skill "I have ADHD", "taste skill", "golang-pro" and "vue-best-practices".

When coding, split the codes into main entrance and several modules to ensure its extendability and maintainability. Write module specification when finish. Organize everything into folders, and do not contain less than 1 files per folder.

The upstream project will be remnawave and emby. Follow their api documents strictly. All api documents are located at ./reference. Make the best of the projects.

No need to test locally but audit all the codes when finished. Ignore all uncommitted git differences.

---

## New features

### Activity
* Replace the "Games Placeholder"
1. Betting with a admin-adjustable winning percentage
2. Everyday sign-in with admin-adjustable TXB awards
3. Lucky draw with admin-adjustable fees and awards (including Coupon, add/deduct balance, expiration date expand, etc.)

### Questionnaire
* Replace the current placeholder

Admin can start a questionnair by setting the external link (google forms/microsoft forms), and setting the TXB reward of it. Once the user choose to fill the form, they will get the link and a validation code. When the admin desided to end the servey, a csv file is required to be uploaded. The program should display the rolls of the csv and let the admin decide which roll is the validation code. The TXB reward is automatically given to the users.

### Emby
* Replace the current placeholder

User needs to pay a preset amount TXB to set up an Emby account.
The account name is default to the username of remnawave.
Transcode, download, remote control, showing on the login screen are prohibited.
If user have a valid linked Emby account, they could adjust their parential rating, passwords, and the media libraries they want to have access to.

### Database Editing

Admins will have full access to the database. Create a database editing GUI to let admins edit user and system data. 

### Coupon

Admins could create coupons, setting the type, limiting the use times, and setting the expire date.
Coupons have the types below:
- recurring offers
- one-time offers
- balance adding
- balance multiplying

### Rollover unused data

When the subscription expires, the remain traffic will be used to calculate how much TXB to give back to the user. Admin could set a maximum TXB rollover and a minimum TXB rollover for each combo.
For example, a user have 100G/200G remaining, and he paid 100TXB for the combo(including extra internal squads), but the maximum TXB rollover is set to be 10TXB, so he will be given 10TXB for rollover.

### Database Clean-up

There should be a maximum 200 audit trails and payment record. Delete the extras when new things are created.


## Adjustments

- Admin could download or recover the database backup on database editing interface
- Delete "Your account at a glance." on the main page of user dashboard
- Add markdown support for combo and squad descriptions
- Show cancel payment button for the interface waiting payment to finish
- Display swithch button instead of "true" or "false" in admin interfaces
- Change xxx per TXB into TXB per xxx (like TXB per CNY)
- Display the catalog using gray color when the internal squad is included
- When the user is in database but unable to locate it in remnawave, the user needs to go through sign-up process again
- Admin should be able to set multiple Payment type for both EZPay and BEPusdt, display to user using secondary menu
- Decrease the use of words and increase the use of icons
- Follow Consumer Psychology design


- Occurs program-wide: When typing TXB values, decimal point is wrongly added (e.g. typed 150 but displayed 1.50)
- Unable to use BEPUsdt, you may want to test the api response locally:
    - Link: https://pay.rna.im
    - API Token: 10EED6C180C11713AD7629353BA06632
- Unable to set included internal squad for catlog
