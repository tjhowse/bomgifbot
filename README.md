# Bomgifbot

This is a bot that pulls images from the web, appends them to a gif and posts
the result to mastodon.


## Hosting

This repo is set up to be run on a fly.io instance to make self-hosting easy. The bring-up process is pretty trivial:

    1. Install git and flyctl
        `sudo apt install git`
        `curl -L https://fly.io/install.sh | sh`
    2. Clone this repo
        `git clone git@github.com:tjhowse/bomgifbot.git`
    3. `flyctl auth signup`
    4. `flyctl init`
    5. Tweak the config in the `env` section of `fly.toml`
    6. Set your secrets
        `flyctl secrets set MASTODON_CLIENT_ID=<your client id> MASTODON_CLIENT_SECRET=<your client secret>`
    7. `flyctl deploy`
    8. To pause the app, run `flyctl scale count 0`
    9. To resume the app, run `flyctl deploy`.

## Configuration

| Setting | Description | Secret | Example | Default |
| --- | --- | --- | --- | --- |
| `MASTODON_SERVER` | The URL of the mastodon server to post to | No | `https://botsin.space` | N/A |
| `MASTODON_CLIENT_ID` | The client ID of the mastodon app to use | Yes | `1234567890` | N/A |
| `MASTODON_CLIENT_SECRET` | The client secret of the mastodon app to use | Yes | `1234567890` | N/A |
| `MASTODON_USER_EMAIL` | The email address of the mastodon account | Yes | `woo@you.com` | N/A |
| `MASTODON_USER_PASSWORD` | The user password of the mastodon account | Yes | `1234567890` | N/A |
| `IMAGE_URL` | The image to regularly download and append to the gif | No | `https://example.com/image.png` | N/A |
| `IMAGE_UPDATE_INTERVAL` | The interval in seconds between downloading the image | No | `300` | `300` |
| `IMAGE_FRAME_COUNT` | The number of frames to keep in the gif | No | `10` | `10` |
| `IMAGE_FRAME_DELAY` | The delay between frames in the gif, in 100ths of a second | No | `1` | `1` |