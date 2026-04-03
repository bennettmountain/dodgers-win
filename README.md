# dodgers-win
A program to send a text whenever Panda Express's DODGERSWIN $7 Panda Plate deal is active. The deal is active the day after the Los Angeles Dodgers win a home game. The text will go out at 5:30 pm PST on those days.

## How can I add myself to the list?
Text "dodgerswin" to XXX-XXX-XXXX to add yourself to the list. You should receive an automated text confirming you properly signed up and instructing you how to remove yourself from the list: by replying "STOP" at any time.


## Tech Stack Overview
How the program works is relatively simple:
- At 5:30 pm every day, a [cron job](https://www.splunk.com/en_us/blog/learn/cron-jobs.html) runs to query the MLB API to obtain results from the previous day's games.
- If the Dodgers won a home game the previous day, the Postgres database (powered by [supabase](https://supabase.com/)) is queried to obtain a list of all active users' phone numbers.
- Using [twilio](https://twilio.com), a text is sent out to all phone numbers present in the database notifying them that the deal is active.
- The server -- which handles all operations described above along with the storing of phone numbers -- is hosted thanks to the free tier of [vercel](https://vercel.com/). Vercel provides a seamless way of running your app in your github repo without the need for containerization.
-- A webhook for the SMS phone number from Twilio handles the sign ups. When a text is sent to the Twilio number, the webhook and `/sms` handler parses the message and takes the correct action of subscribing or unsubscribing from the list.

### Security Concerns
Q: "is my phone number safe?"
A: It's pretty safe. Your phone number is **only** stored in the Postgres database hosted by vercel. It is **not** stored on Github or anywhere on my local machine. If vercel has a data leakage and my database is part of the leakage or hack, then no, your number is not safe and could be leaked. I am not resposible for anyone spam calling or texting you if your number is leaked because of vercel.

### Local Development

#### Database Setup
1. Start Docker Desktop
2. Start supabase with `supabase start`
3. Log in to supabase with `supabase login`
4. Link project with `supabase link --project-ref <project_id>`
5. Push the migration to your remote DB with `supabase db push`
6. View the local database at http://localhost:54323

#### Environment Variables
Create a `.env` file in the project root with the following variables:
```
DB_HOST=<supabase_db_host>
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<supabase_db_password>
DB_NAME=postgres
TWILIO_ACCOUNT_SID=<twilio_account_sid>
TWILIO_AUTH_TOKEN=<twilio_auth_token>
TWILIO_PHONE_NUMBER=<twilio_phone_number>
CONTACT_CARD_URL=<s3_contact_card_url>
DODGERS_WIN_GIF_URL=<s3_dodgers_win_gif_url>
```

### Scripts

#### Check Dodgers Win
Queries the MLB Stats API for yesterday's results and checks if the Dodgers played at home and won. No database or Twilio credentials needed.
```bash
go run cmd/check_dodgers_win/main.go
```

#### Send Text
Sends a text message to a single phone number or to all active subscribers in the remote database. Requires Twilio credentials in `.env`. The "win" text also includes a celebratory gif.
```bash
# Send the win text to a single number
go run cmd/send_text/main.go --phone-number "+11234567890" --text win

# Send the welcome text to a single number
go run cmd/send_text/main.go --phone-number "+11234567890" --text welcome

# Send the win text to all active subscribers
go run cmd/send_text/main.go --all-subscribers --text win
```

#### List Subscribers
Connects to the remote database and lists all rows in the subscribers table. Requires database credentials in `.env`.
```bash
go run cmd/list_subscribers/main.go
```
