# backup-creator

## **Requirements**
script needs to call Salesforce Marketing Cloud APIs in a secure manner; 
no access tokens, credentials or other sensitive data should be kept within build or repository; 
it should be schedulable e.g. once per day; 
it should fetch all updated or new content blocks since last run and copy its content into selected storage i.e. local file storage or cloud-based object storage like Amazon S3 / Google Cloud Storage. 
every run should create a subfolder within a selected storage e.g. “backup_010623


Backup Creator is a tool designed to back up updated or new content blocks from **Salesforce Marketing Cloud**. The program satisfies the following requirements:

---

## **Features**

1. **Secure Salesforce Marketing Cloud API Access:**
   - Access tokens are requested dynamically and are not stored in the repository or build.
   - Credentials (client ID, client secret) are loaded from environment variables.

2. **Task Scheduler:**
   - Supports daily task scheduling using cron expressions.
   - Tracks the last execution time using a local file to fetch only new data.

3. **Fetch Only Updated Data:**
   - Retrieves content blocks that were updated or created since the last run.

4. **Flexible Storage Options:**
   - Local file storage (file system).
   - Cloud-based storage (Amazon S3).
   - Each run creates a subfolder in the format `backup_YYYYMMDD` for data grouping.

5. **Parallel Processing:**
   - Saves content blocks concurrently to improve performance.

---
## How It Works

### **Initialization**
- Set up environment variables: Create a `.env` file in the root directory.
- Configures Salesforce API and storage clients.

### **Task Scheduling**
- Supports flexible scheduling using cron expressionsю
- Determines the last execution time from a file.

### **Data Fetching**
- Fetches only the updated or new content blocks from Salesforce Marketing Cloud.

### **Data Saving**
- Saves content blocks to the selected storage (local or Amazon S3).
- Each run creates a unique folder for the backed-up data.

---

## Possible Improvements (TODO)

### **Test Environment**
- Add unit tests and integration tests. 

### **Error Handling**
- Enhance error handling. 

### **Structured Logging**
- Integrate a logger like `zap` or implement a logger using `slog`.

### **Support for Additional Cloud Storages**
- Implement saving to the Google Cloud Storage or Azure Blob Storage.

### **Add cache to store token**
- Add cache to store token to improve performance (decrease amount of requests to API).

### **Advanced Parallel Processing**
- Use a semaphore to limit the number of concurrently running goroutines.

### **Monitoring and Alerts**
- Add notifications (e.g., Slack or Email) on task success or failure.

 ## RUN
  `docker run -p 8080:8080 --env-file .env backup-creator`

  ## Tests
  `go test -cover -count=1 ./...`
  