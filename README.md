# FFast: FDM-Integrated Link Processor

**FFast** is a specialized Go-based CLI tool designed to streamline downloads from `fuckingfast.co` by integrating directly with **Free Download Manager (FDM)**. It automates the extraction of links from PrivateBin-style URLs, manages download states, and configures FDM's internal settings for a seamless experience.

---

## 🚀 Features

* **Automated Link Extraction:** Uses `chromedp` to scrape and decode links from PrivateBin URLs.
* **FDM Configuration:** Automatically detects FDM's installation, updates its internal SQLite database to set download paths, and restarts the service.
* **Smart Caching:** Remembers your URLs and download progress (mandatory, optional, and selective parts) to resume where you left off.
* **Interactive CLI:** Uses `survey` to provide a user-friendly selection process for optional and selective game parts.

---

## 🛠️ Prerequisites

* **Go:** Version `1.25.0` or higher.
* **OS:** Windows (Required for registry and FDM database pathing).
* **Free Download Manager:** Must be installed on your system.
> [!IMPORTANT]
> **Download FDM:** [Official Website](https://www.freedownloadmanager.org/download.htm)



---

## 🏗️ Building and Installation

To build the executable from the source code, follow these steps:

1. **Clone the Repository:**
```bash
git clone https://github.com/yourusername/FFast.git
cd FFast

```


2. **Install Dependencies:**
```bash
go mod tidy

```


3. **Build the Project:**
```bash
go build -o FFast.exe main.go

```



---

## 📖 How to Use

1. **Launch:** Run the compiled binary:
```bash
.\FFast.exe

```


2. **Input URL:** Enter the PrivateBin/Fucking-Fast URL when prompted.
3. **Configure Path:** If it's your first time, the tool will ask for your preferred FDM download directory. It will then update FDM and restart it in the background.
4. **Select Parts:** Use the arrow keys and spacebar to select **Optional** or **Selective** files you wish to download.
5. **Sit Back:** The tool will feed the links to FDM one by one and track completion.

---

## 📁 Project Structure

| Directory/File | Description |
| --- | --- |
| `ffast/` | Core package containing downloader logic. |
| `ffast/cacheCode/` | Handles SHA-1 hashing and file-based state caching. |
| `ffast/fdmConfig/` | Windows Registry and FDM SQLite database configuration. |
| `main.go` | Entry point for the CLI application. |
| `cache/` | (Ignored) Stores session data and download states. |

---

## 🤝 How to Commit

When contributing to this project, please follow these steps to ensure a clean history:

1. **Sync your local branch:**
```bash
git pull origin main

```


2. **Stage your changes:**
```bash
git add .

```


3. **Commit with a descriptive message:**
* *Example:* `git commit -m "feat: add support for new link patterns"`


4. **Push to your branch:**
```bash
git push origin your-branch-name

```



---

**Note:** This tool is intended for personal automation and efficiency. Ensure you have the rights to the content you are downloading.

Would you like me to create a `.github/workflows` file to automate the build process for you?