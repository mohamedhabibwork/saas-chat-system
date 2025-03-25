# File Handling Documentation

## Overview

The file handling system allows users to upload, download, and manage files within their subscription limits. The system supports both local storage and Amazon S3 for file storage.

## Features

- File upload with subscription-based limits
- File download with access control
- File deletion
- File listing
- Storage usage tracking
- Support for local and S3 storage
- Configurable file type restrictions

## Configuration

The file handling system is configured through `config/storage.json`. Key configuration options include:

- Storage type (local or S3)
- Upload paths
- S3 credentials and settings
- File size limits
- Allowed file extensions and MIME types

### Local Storage Configuration

```json
{
    "storage": {
        "type": "local",
        "local": {
            "upload_path": "./uploads",
            "temp_path": "./temp"
        }
    }
}
```

### S3 Storage Configuration

```json
{
    "storage": {
        "type": "s3",
        "s3": {
            "region": "us-west-2",
            "bucket": "your-bucket-name",
            "access_key_id": "your-access-key",
            "secret_access_key": "your-secret-key"
        }
    }
}
```

## API Endpoints

### Upload File

```http
POST /api/files/upload
Content-Type: multipart/form-data

file: [binary]
```

Response:
```json
{
    "message": "File uploaded successfully",
    "data": {
        "id": 1,
        "filename": "example.pdf",
        "url": "/files/uploads/example.pdf",
        "size": 1024,
        "content_type": "application/pdf",
        "created_at": "2024-03-20T10:30:00Z"
    }
}
```

### Download File

```http
GET /api/files/download?id=1
```

Response: Binary file content with appropriate headers

### Delete File

```http
DELETE /api/files/delete?id=1
```

Response:
```json
{
    "message": "File deleted successfully"
}
```

### List Files

```http
GET /api/files/list
```

Response:
```json
{
    "message": "Files retrieved successfully",
    "data": [
        {
            "id": 1,
            "filename": "example.pdf",
            "url": "/files/uploads/example.pdf",
            "size": 1024,
            "content_type": "application/pdf",
            "created_at": "2024-03-20T10:30:00Z"
        }
    ]
}
```

## Subscription Limits

File handling is subject to the following subscription-based limits:

- Maximum file size
- Total storage space
- Number of files per day
- Allowed file types

These limits are defined in the subscription plan's `limits` field:

```json
{
    "limits": {
        "max_storage_gb": 10,
        "max_file_size_mb": 100,
        "max_files_per_day": 50
    }
}
```

## Usage Tracking

The system tracks the following metrics for each subscription:

- Total storage used
- Number of files uploaded
- Daily upload count

This information is used to enforce subscription limits and is available through the subscription usage API.

## Security Considerations

1. File access is restricted to the file owner
2. File types are validated before upload
3. File size is checked against subscription limits
4. Storage space is monitored
5. S3 credentials are stored securely
6. Temporary files are cleaned up
7. File paths are sanitized to prevent directory traversal

## Error Handling

The system provides detailed error messages for common issues:

- File size exceeds limit
- Storage quota exceeded
- Invalid file type
- Unauthorized access
- File not found
- Upload/download failures

## Best Practices

1. Use environment variables for sensitive configuration
2. Regularly monitor storage usage
3. Implement file cleanup policies
4. Use CDN for frequently accessed files
5. Implement rate limiting for file operations
6. Regular security audits
7. Backup important files

## Development Setup

1. Create necessary directories:
   ```bash
   mkdir -p uploads temp
   chmod 755 uploads temp
   ```

2. Configure storage settings in `config/storage.json`

3. Set up environment variables:
   ```bash
   export AWS_ACCESS_KEY_ID=your-access-key
   export AWS_SECRET_ACCESS_KEY=your-secret-key
   ```

4. Update database schema:
   ```sql
   -- Run the migration scripts to create the files table
   ```

5. Test file operations:
   ```bash
   # Upload test
   curl -F "file=@test.pdf" http://localhost:8080/api/files/upload
   
   # Download test
   curl http://localhost:8080/api/files/download?id=1
   
   # List files
   curl http://localhost:8080/api/files/list
   ```
``` 