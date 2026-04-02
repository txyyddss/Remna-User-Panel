# Goal
I want you to develop a carpool user panel + telegrm bot with:
- Totally splited frontend and backend
- Golang for backend (Dedicated apis and sdks)
- Vue.js for frontend
- Github actions for backend/frontend compiling and issue/pull request auto validating
- Sqlite for backend database
- Full integration of telegram mini app
- Shows refuse page when non-tg access is detected
- Adjusts webpage layout dynamically based on size of webpage
- Integrated management functions for admins or using api tokens
- Jellyfin account creation and management for users
- VPN carpool using remnawave
- Integrated payment using BEPusdt
- Functions of remnawave and jellyfin should be splited but shares the same credit system
- Auto-login for users based on telegram id

Read files under ./Development Guide for detailed instructions
All files under ./Remnawave , ./EZPay , ./Jellyfin and ./BEPusdt are the files of REST APIs of upstream projects you may need
After you finish writing all the codes, generate README.md of the project and detailed usage, techinical documents, api docs and sdk docs under docs folder

Here are the things for you to understand the response format of upstream projects. DO NOT perform actions to existing users/datas. You are allowed perform adding actions but remember to delete it afterwards.

Remnawave:
- URL: https://panel.1391399.xyz/
- Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiMzEwMDhkNWYtZWI2Ni00NDU2LWIyYWQtMTU1ZDlhZTdiYzk0IiwidXNlcm5hbWUiOm51bGwsInJvbGUiOiJBUEkiLCJpYXQiOjE3NzUwNTQ5ODMsImV4cCI6MTA0MTQ5Njg1ODN9.X-7OcfRjOHtZHbzt4A97PM7xRfmILDPXnJRxz9eb11A

Jelyfin:
- URL: http://145.239.244.62:28096/
- Token: 467eafea7c8b4f98be7b1a4d5eed42e5

BEPusdt:
- URL: https://pay.rna.im
- Token: 10EED6C180C11713AD7629353BA06632

EZPay:
- URL: https://eb9a928f.aodww.cn
- 商户ID: 1202
- 商户密钥: 66C635aAc640vc6tzAhCM8c3EVe35WS0

# Payment
All of the payment actions should be using one of BEPusdt (Display as USDT) and EZPay (Display as 易支付). User should select one interface and select specialized payment method of each interface. User could choose whether or not using credits to discount the bill. Skip payment when the value is 0. Return error when the value <0.

# The Credit system
The credit should be called as TXB. Credits could be added by Signup, consuming money, sending group messages, and buying credits. The credit should be stored in 2 decimal places and can be below zero. Check if the user have enough TXB to perform credit adding actions. All of the record of credit change should be in the database for 30 days for user to check.
## Signup
User should be signing up using /signup. User could not use the function when his credit is below zero. Only one signup should be performed per day. Send a message showing the value after signing up. Messages should be automatically deleted after 10 seconds when the user receives <1 TXB for signing up. The default random value of signing up is 0~10(2d.p.) with low possibility to get high value and higher possibility to get low value. 
## Consuming money
TXBs should be added once payment. When TXB is used for discount of the bill, the program should not add credit for that bill. The default currency of RMB->TXB is 10 (configurable), which means the user gets 300 TXB when consumed 30 RMB.
## Using TXB for discount
TXB could be used to discount a bill. Default value is 100TXB to 1RMB(configurable), which means if user have 400TXB and he is paying a 30RMB bill, he should only pay 26RMB and 400TXB is deducted from his account. There should only be integer of RMB discount, for if the user have 450TXBs, he could only use 400TXB(4RMB) for discount.
## Betting
User could send /bet <value> for betting. When betting triggers, <value> should be deducted from the user. The default range of credit user can get is -<value>x3~<value>x2(can be 2d.p., configurable), for if user bet for 30TXB, he may get -90TXB~60TXB. The higher the value, the higher the possibility to get lower value. The lower balance of the user, the higher possibility to get higher value, vice versa. The more times of user betting in one month, the lower possibility to get value >0. Message should be sent showing the credit added/deducted and the user's current balance.
## Sending group messages
Once group message is sent, the program should add it into database. Automatically exclude bot messages, private messages, and messages with media files.
Once the message reaches 20 or the customized number, it should be send to OpenAI compatible AI and evaluate its value.
- The token, base_url, and model should be configured in config.json or edited using api
- It should be an optional module
- Example of deepseek request
```
curl https://api.deepseek.com/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${DEEPSEEK_API_KEY}" \
  -d '{
        "model": "deepseek-chat",
        "messages": [
          {"role": "system", "content": "You are a helpful assistant."},
          {"role": "user", "content": "Hello!"}
        ],
        "stream": false
      }'
```
- The accepted credit range should be costomized and have 1 decimal place, default -2~3.
- Write a prompt for AI
- Add an ID for each message, and let AI only return the ID and value of each message using json format.
- Add the credit to database of the user and delete the messages from the database
- It should be high credit for valuable sharing, medium credits for answering questions, low credit for questioning, and deduct credit for attacking others, sending ads, useless group commands, or other inappropriate messages
- Send leaderboard of top 5 users per 100 message(configurable), showing the credit change data of that 100 message.
# Telegram group commands
Users can use /sub to show their sub status or reply a message using /sub showing the sub of sender. When requested, the program should reply a message like the format below, do not include the descriptions:
📊 我的订阅 🟢  -> Green-Active, Yellow-Limited, Red-Expired/Disabled

⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️ 0% | 0 B/500 GB -> progress bar of bandwith usage, changes colour for different percent of usage
📅 剩余 233 天 · 上车 132 天 -> Days before expire and days from creation of that user
[▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░] 1.51 GB -> Bar showing percentage of each Node usage of the user

▓ LY-UK (GB) - 1.27 GB (83.7%)
░ AK-HK (HK) - 251.45 MB (16.2%)
█ BR-DE (DE) - 308.57 KB (0.0%)
▒ WW-HK (HK) - 130.95 KB (0.0%)
▇ LY-JP (JP) - 11.06 KB (0.0%)

-> Top 5 Node usage and percentage


# Configuring and API general requirements
- All of the settings should be fully customizeable through api/web and local config.json
- Some of the settings (like token,etc.) should not be edited through api/web
- When receive config api calls, automatically edit config.json and do fast reload
- Backup config.json and database per day
- Clean up outdated datas when backup finishes
- Automatically delete the backup when it exceeds 10 days
- Use Encryption for all the api call parameters
- Manage permission of each token precisely
- Check if the token is valid for every api call
- There should be fundemental and optional parameters for api calls
- The structure of api response should be stable and well structurized
- You should check if all the parameters are valid in every single request
- You should not directly pass the json content from upstream to user
- Prioritize data gets from upstream projects but not local
- Do not store unecessary data

# Combo
## Admin-Creating combos
A combo should include:
- Internal Squad (Remnawave) -> Add an api to get internal squads first
- Traffic in GB (Remnawave) -> Auto convert GB into bytes
- trafficLimitStrategy (Remnawave)
- Payment cycle (Remnawave)
- Name of the combo
- Description of combo (supports Markdown)
- Price of combo (Payment) -> Should be in RMB
- Price of traffic reset (Payment) -> Should be in RMB

It should be stored into the database.
Once adding, the program should generate an exclusive uuid for it for identification.
## Purchasing Combo/binding subscription link
You may refer to /Reference/remna.js
## Renewal
User can choose to renew the same combo or change to another combo.
# Database editing
You should provide interfaces and apis for admins to edit the database directly.
# Jellyfin
Default price of the Jellyfin account is 2RMB/month. The account should be created once purchase and deleted once expire. User can adjust his MaxParentalRating anytime by dragging the slider, within value of 0~22. The account should not be displayed publically. Transcoding should be disabled during user creation, and should not be activated any time. When user have an active account, he could perform the following actions:
- Authorizing Quick Connect
- Changing password
- Getting his device status

# Changing IP
You may refer to /Reference/changeip.js

# Displaying Info
You may add a page to display the infos of the user using graphs. Display the following infos:
- Bandwith Status (Pie Chart)
- Hwid Devices (List)
- IP List (List)
- Subscription Request History of recent 24 hours (List)
- trafficLimitStrategy
- Traffic used/Traffic Limit


# Editing External Squad
User could edit his external squad freely. The program should show user a list of external squad from remnawave for user to choose if he has a active subscription.

# Sub center
You should display user's subscription info and connection keys in a page if he has a active subscription.

# Github Actions
Once detected git push of program files(except unrelated files like .md,etc.), automatically compile and publish release.
The action should publish release of backend and showing resent pushes in changelog.
Compile the static web files to "web" branch and show instructions about how to publish it to cloudflare pages.

# ISSUE template
You should generage issue form template that clearly shows instructions reporting bugs/requesting features. Add action files to run tests for pull requests.

