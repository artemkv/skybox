# Skybox

Skybox is my personal replacement for Cloud storage services like Google Drive or Dropbox.

Essentially, it is a backup to a personal S3 bucket.

## Why not Google Drive

- The main premise of such systems, namely mirroring files across devices, is an overkill;
- Running these tools continuously eats into resources;
- The free space is limited, subscriptions have weird structure, not control over the cost;
- You are at the mercy of the vendor, and useful features sometimes are hidden behind paywall;
- I don't trust anyone with my files. I don't want vendors to scan my files.

## My personal main use cases

- Primary (home) laptop folder continuous backup;
- Secondary (work) laptop folder continuous backup;
- Download and open any file from backup at any time from a mobile app;
- Files that are modified often are short;
- Files that are long are almost never modified.

## Detailed requirements

- Each device that is backing up is identified using `deviceId` (unique string, user provided);
- Devices are virtual, you can technically back up multiple folders of the same physical device as devices, using separate instances of a tool;
- Every device is associated with one single local folder, the backup saves every file inside that folder to S3 bucket;
- All files that belong to a device are stored in a separate "subfolder" per device in S3;
- You control the S3 bucket to back up the device; you may want to use different storage classes for different devices;
- There is no backend;
- Running a tool makes the backup and exits. Use third party tools to schedule the tool to run periodically;
- Running a tool in restore mode restores every file from S3 to the local folder;
- Restore is intended to be run once, to prime the new device, it is not a mirroring tool;
- File content is stored in S3 encrypted (ChaCha20);
- Encryption is done without Poly1305 to maintain 1:1 byte match with original file, in case I want to stream part of the content and decrypt chunks without reading the complete file (e.g. video);
- Metadata (for now) is stored as plain text (JSON), includes file paths (relative to folder), size and hash (blake3);
- Files with the same content are stored in S3 once (de-duped by hash);
- Error reading local file never affects files stored in S3. In other words, I never delete the file from S3 if I am not sure. This means orphans may be present temporarily, until the errors are resolved.

## Config

Provide values as environment variables or using `.env` file.

```
SKYBOX_FOLDER=<path to local folder>
SKYBOX_BUCKET=<bucket name>
SKYBOX_DEVICEID=<unique device id>
SKYBOX_SECRET=<secret, long string of text>

AWS_REGION=<us-east-1>
AWS_ACCESS_KEY_ID=<access_key>
AWS_SECRET_ACCESS_KEY=<secret_key>
```

## Bucket settings

### Resources

- IAM user with associated Access key
- Regular, general purpose AWS S3 bucket with no public access

### Bucket Policy

Assuming bucket name "mybucketname" and user "skyboxuser",

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::XXXXXXXXXXXX:user/skyboxuser"
            },
            "Action": [
                "s3:ListBucket",
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::mybucketname",
                "arn:aws:s3:::mybucketname/*"
            ]
        }
    ]
}
```

### Cross-origin resource sharing (CORS)

Required to use with Skybox app

```
[
    {
        "AllowedHeaders": [
            "*"
        ],
        "AllowedMethods": [
            "GET",
            "HEAD"
        ],
        "AllowedOrigins": [
            "https://localhost",
            "http://localhost:5173"
        ],
        "ExposeHeaders": [
            "ETag",
            "Content-Length",
            "Content-Type",
            "x-amz-meta-filekey",
            "x-amz-meta-filekeynonce",
            "x-amz-meta-nonce"
        ]
    }
]
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
