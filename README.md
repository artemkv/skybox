# Skybox

Skybox is my personal replacement for Cloud storage services like Google Drive or Dropbox.

Essentially, it is a backup to personal S3 bucket.

## Why

- The main premise of such systems, namely mirroring across devices, is an overkill;
- The free space is limited and subscriptions are costly;
- The vendor lock-on;
- Need to use multiple tools from different vendors;
- I don't trust anyone with my files. I don't want vendors scan my files.

## My personal main use cases

- Home laptop folder continuous backup;
- Work laptop folder continuous backup;
- Download any file from backup to any device, including mobile;
- Files that are modified often are short;
- Files that are long are almost never modified.

## Requirements

- Each device is identified using `deviceId` (unique string);
- All files that belong to that device are stored in a separate "subfolder" in S3;
- Backup works on one folder per device, saves every file inside that folder to S3;
- Backup is intended to run periodically, use third party tools to schedule;
- Restore works on one folder per device, restores every file from S3 to that folder;
- Restore is intended to be run once, to prime the new device;
- File content is stored in S3 encrypted (ChaCha20);
- Encryption is done without Poly1305 to maintain 1:1 byte match with original file, in case I want to stream part of the content and decrypt at the same time (e.g. video);
- Metadata is stored as plain text (JSON), includes file paths (relative to folder), size and hash (blake3);
- Files with the same content are stored in S3 once;
- Error reading local file never affects files stored in S3;
- Orphans should be avoided, but may happen.

## Config

```
SKYBOX_FOLDER=<path to folder>
SKYBOX_BUCKET=<bucket>
SKYBOX_DEVICEID=<unique device id>
SKYBOX_SECRET=<secret>

AWS_REGION=<us-east-1>
AWS_ACCESS_KEY_ID=<access_key>
AWS_SECRET_ACCESS_KEY=<secret_key>
```

## Running Skybox

Backup

```
skybox
skybox backup
```

Restore

```
skybox restore
```
