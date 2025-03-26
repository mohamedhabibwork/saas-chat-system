# Subscription Plans and File Usage Limits

## Overview
This document describes the subscription plans available in the system and their associated file usage limits and features.

## Subscription Plans

### Free Plan
- **Storage**: 1 GB
- **Files**: 100 files
- **Daily Uploads**: 10 files
- **Max File Size**: 10 MB
- **Features**:
  - Basic file sharing
  - Standard file types
  - No video chat
  - No screen sharing
  - No bot integration

### Basic Plan
- **Storage**: 5 GB
- **Files**: 1,000 files
- **Daily Uploads**: 50 files
- **Max File Size**: 50 MB
- **Features**:
  - Enhanced file sharing
  - Extended file types
  - Video chat
  - No screen sharing
  - No bot integration

### Professional Plan
- **Storage**: 20 GB
- **Files**: 10,000 files
- **Daily Uploads**: 200 files
- **Max File Size**: 100 MB
- **Features**:
  - Advanced file sharing
  - Professional file types
  - Video chat
  - Screen sharing
  - Bot integration

### Enterprise Plan
- **Storage**: 100 GB
- **Files**: 100,000 files
- **Daily Uploads**: 1,000 files
- **Max File Size**: 500 MB
- **Features**:
  - Unlimited file sharing
  - All file types
  - Video chat
  - Screen sharing
  - Bot integration
  - Custom bot development
  - Priority support
  - Dedicated server

## File Type Restrictions

### Free Plan
- Images: JPG, JPEG, PNG, GIF
- Documents: PDF, DOC, DOCX, TXT
- Spreadsheets: CSV, XLS, XLSX

### Basic Plan
- All Free Plan types
- Presentations: PPT, PPTX
- Archives: ZIP, RAR
- Media: MP3, MP4, AVI, MOV

### Professional Plan
- All Basic Plan types
- Design files: PSD, AI, EPS
- Vector graphics: SVG
- Web media: WEBP, WEBM
- Additional video: MKV, FLV, WMV

### Enterprise Plan
- All Professional Plan types
- System files: ISO, DMG, EXE, MSI
- Mobile apps: APK, IPA
- Package files: DEB, RPM

## Usage Tracking

### Storage Usage
- Total storage used is tracked in bytes
- Storage usage is updated in real-time
- Historical storage usage is available

### File Count
- Total number of files is tracked
- Daily file upload count is maintained
- File count limits are enforced

### Daily Uploads
- Daily upload counter resets at midnight
- Upload limits are enforced per day
- Historical upload data is available

## Feature Access

### Video Chat
- Available in Basic, Professional, and Enterprise plans
- Quality settings based on plan level
- Participant limits vary by plan

### Screen Sharing
- Available in Professional and Enterprise plans
- Quality settings based on plan level
- Multiple screen support in Enterprise plan

### File Sharing
- Available in all plans
- Sharing limits vary by plan
- Advanced sharing features in higher plans

### Bot Integration
- Available in Professional and Enterprise plans
- Custom bot development in Enterprise plan
- API access levels vary by plan

## Usage Monitoring

### Dashboard
- Real-time usage statistics
- Storage usage visualization
- File type distribution
- Upload history

### Alerts
- Storage limit warnings
- Daily upload limit notifications
- Feature usage alerts
- Plan upgrade suggestions

### Reports
- Monthly usage reports
- Storage usage trends
- File type analysis
- Feature usage statistics

## Plan Management

### Upgrades
- Seamless plan upgrades
- Prorated billing
- Feature access immediate
- Storage limits adjusted

### Downgrades
- Storage limit enforcement
- Feature access removal
- File type restrictions
- Usage monitoring

### Cancellations
- Data retention policy
- Export options
- Backup options
- Account closure process

## Technical Implementation

### Database Schema
```sql
CREATE TABLE subscription_plans (
    name VARCHAR(50) PRIMARY KEY,
    max_storage BIGINT NOT NULL,
    max_files INTEGER NOT NULL,
    max_daily_uploads INTEGER NOT NULL,
    max_file_size BIGINT NOT NULL,
    allowed_extensions TEXT[] NOT NULL,
    allowed_mime_types TEXT[] NOT NULL,
    features JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    tenant_id INTEGER NOT NULL REFERENCES tenants(id),
    plan VARCHAR(50) NOT NULL REFERENCES subscription_plans(name),
    status VARCHAR(20) NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    storage_used BIGINT NOT NULL DEFAULT 0,
    files_uploaded INTEGER NOT NULL DEFAULT 0,
    daily_uploads INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

### Configuration
Subscription plans are configured in `config/subscription_plans.json`:
```json
{
    "plans": {
        "free": {
            "name": "Free",
            "max_storage": 1073741824,
            "max_files": 100,
            "max_daily_uploads": 10,
            "max_file_size": 10485760,
            "allowed_extensions": [...],
            "allowed_mime_types": [...],
            "features": {...}
        },
        ...
    }
}
```

## API Endpoints

### Subscription Management
- `GET /api/subscriptions` - List user's subscriptions
- `POST /api/subscriptions` - Create new subscription
- `PUT /api/subscriptions/:id` - Update subscription
- `DELETE /api/subscriptions/:id` - Cancel subscription

### Usage Monitoring
- `GET /api/subscriptions/:id/usage` - Get subscription usage
- `GET /api/subscriptions/:id/history` - Get usage history
- `GET /api/subscriptions/:id/storage` - Get storage usage
- `GET /api/subscriptions/:id/files` - Get file usage

### Plan Management
- `GET /api/plans` - List available plans
- `GET /api/plans/:name` - Get plan details
- `POST /api/plans/:name/upgrade` - Upgrade to plan
- `POST /api/plans/:name/downgrade` - Downgrade to plan

## Security Considerations

### Access Control
- Role-based access control
- Feature-level permissions
- API access restrictions
- Rate limiting

### Data Protection
- Encrypted storage
- Secure file transfer
- Access logging
- Audit trails

### Compliance
- Data retention policies
- Privacy regulations
- Security standards
- Usage monitoring

## Support

### Technical Support
- Email support
- Documentation
- API reference
- Troubleshooting guides

### Account Management
- Billing support
- Plan changes
- Usage questions
- Feature requests

### Emergency Support
- 24/7 availability
- Priority response
- Escalation process
- Backup assistance

## Feature Comparison

| Feature | Free | Basic | Pro | Enterprise |
|---------|------|-------|-----|------------|
| **Chat Features** | | | | |
| Real-time messaging | ✓ | ✓ | ✓ | ✓ |
| Group chats | ✓ | ✓ | ✓ | ✓ |
| File sharing | 10MB | 50MB | 100MB | Unlimited |
| Message history | 7 days | 30 days | 1 year | Unlimited |
| **Forum Features** | | | | |
| Forum categories | 1 | 5 | 20 | Unlimited |
| Topics per category | 10 | 50 | 200 | Unlimited |
| Posts per topic | 50 | 200 | 1000 | Unlimited |
| File attachments | 5MB | 20MB | 50MB | Unlimited |
| Content moderation | Basic | Basic | Advanced | Custom |
| **Notification Features** | | | | |
| Email notifications | ✓ | ✓ | ✓ | ✓ |
| Push notifications | - | ✓ | ✓ | ✓ |
| Custom notification templates | - | - | ✓ | ✓ |
| Webhook integration | - | - | 5 | Unlimited |
| Notification preferences | Basic | Basic | Advanced | Custom |
| **Support Features** | | | | |
| Email support | ✓ | ✓ | ✓ | ✓ |
| Priority support | - | - | ✓ | ✓ |
| 24/7 support | - | - | - | ✓ |
| Dedicated account manager | - | - | - | ✓ |
| **Security Features** | | | | |
| End-to-end encryption | ✓ | ✓ | ✓ | ✓ |
| Two-factor authentication | - | ✓ | ✓ | ✓ |
| SSO integration | - | - | ✓ | ✓ |
| Custom security policies | - | - | - | ✓ |

## Pricing

| Plan | Monthly | Annual (20% off) |
|------|---------|------------------|
| Free | $0 | $0 |
| Basic | $29 | $278 |
| Pro | $99 | $950 |
| Enterprise | Custom | Custom |

## Feature Details

### Forum Features

1. **Categories**
   - Organize topics by category
   - Set category permissions
   - Custom category ordering
   - Category-specific moderators

2. **Topics**
   - Create and manage topics
   - Pin important topics
   - Lock topics for announcements
   - Topic-specific permissions

3. **Posts**
   - Rich text formatting
   - File attachments
   - Code blocks with syntax highlighting
   - Post editing and deletion

4. **Moderation**
   - Content filtering
   - Spam detection
   - User reporting
   - Automated moderation rules

### Notification Features

1. **Email Notifications**
   - HTML email templates
   - Custom branding
   - Multiple languages
   - Digest options

2. **Push Notifications**
   - Real-time updates
   - Custom notification sounds
   - Notification grouping
   - Deep linking

3. **Webhook Integration**
   - Custom webhook endpoints
   - Event filtering
   - Retry mechanisms
   - Webhook monitoring

4. **Notification Preferences**
   - Per-user settings
   - Per-category settings
   - Quiet hours
   - Notification frequency 