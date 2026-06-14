# 🗂️ What is TubeCast
- Converts YouTube channels or videos into personal podcast feeds.
- Lets you follow these feeds in any podcast app (like Apple Podcasts, Castbox, or Pocket Casts) simply by adding the feed URL.
- Uploads everything to the Internet Archive for free, cloud-based hosting.
- Feeds are always up-to-date—new episodes appear in your podcast app whenever you sync.
- Listen on any device, even with your screen off—just like a real podcast.
- Share your feeds easily with friends and family.

## Demo
- Paste this show link on your podcast app: `https://archive.org/download/fitrahhaque_tubecast/bloop.xml`
- You will find it in `Shows` (for Apple Podcasts) list if it does not appear right away.

<div align="center">
<img src="donotopen/podcast-url.png" alt="Apple Podcast Follow" width="277" height="600"/><img src="donotopen/podcast-preview.png" alt="Apple Podcast Preview" width="277" height="600"/><img src="donotopen/castbox-preview.png" alt="Castbox for Android" width="277" height="600"/> 
</div> 

## Quick Start

The [`tubecast-scripts`](https://drive.google.com/file/d/1aAkkPFOyZulgHiwneI6GSi2wiICHdFd7/view?usp=sharing) package runs TubeCast in Docker and opens an interactive terminal user interface (TUI). All show management is done from this interface.

### Requirements

- Docker Desktop on macOS or Windows, or Docker Engine with Compose on Linux.
- An [Internet Archive account](https://archive.org/account/signup).
- A terminal. On Windows, use WSL or another Bash-compatible terminal for the included `.sh` scripts.

### Package Contents

| File | Purpose |
| --- | --- |
| `example.txt` | Environment template. Copy it to `.env`. |
| `docker-compose.yml` | Container, storage, and Internet Archive configuration mounts. |
| `init.sh` | Pulls the TubeCast image and downloads the Internet Archive CLI. |
| `run.sh` | Updates the image when possible and starts the TubeCast TUI. |
| `tubecast/cover/` | Put show cover images here before creating a show. |

### First-Time Setup

Unzip the package and enter its directory:

```bash
cd tubecast-scripts
```

Create your environment file:

```bash
cp example.txt .env
```

Open `.env` and set a unique username:

```dotenv
USERNAME="yourname"
ARCHIVE="Yes"
```

The username becomes part of your Internet Archive item name: `yourname_tubecast`. Use lowercase letters without spaces.

Make the scripts executable and initialize the package:

```bash
chmod +x init.sh run.sh
./init.sh
```

Configure the Internet Archive CLI using your verified account:

```bash
./ia configure
```

Your password is not displayed while you type it. The resulting configuration is stored at `~/.config/internetarchive/ia.ini` and mounted into the TubeCast container automatically.

### Start TubeCast

```bash
./run.sh
```

The first launch may take a moment while Docker downloads the image.

## TUI Tutorial

### Keyboard Controls

| Key | Action |
| --- | --- |
| `Up` / `Down` | Move through menu items and form fields. |
| `Enter` | Open the selected item, press a button, or confirm an action. |
| `Tab` | Move between form fields and buttons. |
| `Esc` | Return to the main menu from a form. |
| `b` | Go back from a list that shows a Back option. |
| `q` | Quit from the main menu. |

The main menu also shows shortcuts: `v` for shows, `c` to create a show, `s` to subscribe, `a` to sync, `i` to add episodes, `d` to delete a show, and `q` to quit.

For the Docker package, change `USERNAME` or `ARCHIVE` by editing the host `.env` file and restarting TubeCast. The TUI's **set env** screen is intended for native runs and does not persist after a temporary Docker container exits.

### 1. Create a Show

Before opening the TUI, put a `.png` or `.webp` cover image in `tubecast/cover/`. Podcast artwork should be square and between 1400x1400 and 3000x3000 pixels.

Example:

```text
tubecast/cover/medicine.png
```

In the TUI:

1. Select **create a show**.
2. Enter the show title, for example `Medicine`.
3. Enter a description.
4. Enter the cover filename only, for example `medicine.png`.
5. Select **save**.
6. Select **COPY SHOW LINK** to copy the podcast feed URL.

Use the exact same show title in later actions. Show titles are case-sensitive.

### 2. Subscribe to a YouTube Channel

Subscribing adds the latest three videos from a channel and remembers the channel for future syncs.

1. Select **subscribe**.
2. Enter an existing show title.
3. Enter the YouTube channel handle, such as `ThePrimeTimeagen` or `@ThePrimeTimeagen`.
4. Select **Add** and wait for the downloads and uploads to finish.
5. Copy the show link from the success message.

### 3. Add Individual Episodes

1. Select **add episodes**.
2. Enter an existing show title.
3. Enter one or more YouTube video URLs. Separate multiple URLs with commas.
4. Select **Add** and wait for processing to finish.

Example input:

```text
https://youtu.be/video1,https://youtu.be/video2
```

TubeCast downloads the audio and thumbnail, uploads them to Internet Archive, and updates the podcast feed.

### 4. Sync Subscribed Channels

Select **sync** from the main menu. TubeCast checks every channel subscribed to by every show and adds new videos from each channel's latest three results. Videos already in a show are skipped.

### 5. Browse Shows and Copy a Feed URL

1. Select **shows**.
2. Select a show to view its episodes.
3. Select **Copy link to clipboard** to copy the feed URL.
4. Use your podcast app's **Follow a Show by URL** or equivalent option and paste the URL.

You only need to follow the URL once. Future syncs update the same feed.

### 6. Remove an Episode

1. Open **shows** and select a show.
2. Select the episode you want to remove.
3. Choose **YES** in the confirmation dialog.

This removes the episode files from Internet Archive and updates the feed.

### 7. Delete a Show

Select **delete a show**, then select the show to delete.

**Warning:** selecting a show in this menu starts deletion immediately. It removes the local show data and its files from Internet Archive.

### Stop TubeCast

Select **quit** or press `q` from the main menu. Docker removes the temporary container, while your shows, feeds, and cover files remain in the `tubecast-scripts` directory.

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `docker: command not found` | Install and start Docker Desktop, or install Docker Engine and Compose. |
| Internet Archive configuration not found | Run `./ia configure` and confirm `~/.config/internetarchive/ia.ini` exists. |
| YouTube format or signature errors | Run `docker compose pull` to download the latest TubeCast image. |
| Permission denied when running a script | Run `chmod +x init.sh run.sh`. |
| A show is not visible in a podcast app yet | Wait a few minutes for Internet Archive and the podcast app to refresh their caches. |

---

## 📄 License
