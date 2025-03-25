# Data Initialization Script

This script initializes the basic data required for the system to function properly. It creates default roles, permissions, subscription plans, a default tenant, an admin user, and default channels.

## Prerequisites

- PostgreSQL database installed and running
- Go 1.16 or later installed
- Database connection details (host, port, username, password, database name)

## Configuration

Before running the script, you need to:

1. Generate a secure password hash for the admin user:
   ```bash
   go run generate_password.go -password "your_secure_password"
   ```
   This will output a bcrypt hash that you can use in the initialization script.

2. Update the database connection string in `init_data.go`:
   ```go
   db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/dbname?sslmode=disable")
   ```

   Replace the following placeholders with your actual database credentials:
   - `user`: Database username
   - `password`: Database password
   - `localhost`: Database host
   - `5432`: Database port
   - `dbname`: Database name

3. Update the admin user's hashed password in `init_data.go`:
   ```go
   hashedPassword := "your_generated_hash_here" // Replace with the hash from step 1
   ```

## What Gets Created

The script creates the following:

### Roles
- Admin: System administrator with full access
- User: Regular user with basic access
- Moderator: Channel moderator with limited admin access
- Bot: Bot user with limited access

### Permissions
- create_channel
- delete_channel
- manage_users
- manage_roles
- manage_subscriptions
- view_analytics
- manage_bots
- upload_files
- download_files
- share_files

### Subscription Plans
1. Free Plan
   - Price: $0/month
   - Features: Basic file sharing
   - Limits: 1GB storage, 100 files, 10 daily uploads, 10MB max file size

2. Basic Plan
   - Price: $9.99/month
   - Features: Video chat, file sharing
   - Limits: 5GB storage, 1,000 files, 50 daily uploads, 50MB max file size

3. Pro Plan
   - Price: $29.99/month
   - Features: Video chat, screen sharing, file sharing, bot integration
   - Limits: 20GB storage, 10,000 files, 200 daily uploads, 100MB max file size

4. Enterprise Plan
   - Price: $99.99/month
   - Features: All features including custom bot development
   - Limits: 100GB storage, 100,000 files, 1,000 daily uploads, 500MB max file size

### Default Tenant
- Name: "Default Tenant"
- Status: active

### Admin User
- Username: admin
- Email: admin@example.com
- Role: admin
- Status: active
- Subscription: Enterprise plan (1 year)

### Default Channels
1. general
   - Type: public
   - Description: General discussion channel

2. announcements
   - Type: public
   - Description: System announcements

3. support
   - Type: public
   - Description: Customer support channel

4. admin
   - Type: private
   - Description: Administrative channel

## Running the Script

1. Navigate to the scripts directory:
   ```bash
   cd scripts
   ```

2. Generate the password hash:
   ```bash
   go run generate_password.go -password "your_secure_password"
   ```

3. Update the `init_data.go` file with the generated hash and database credentials

4. Run the initialization script:
   ```bash
   go run init_data.go
   ```

The script will create all the necessary data and print a success message when completed. If any errors occur, they will be displayed with detailed information about what went wrong.

## Error Handling

The script includes error handling for:
- Database connection issues
- Duplicate entries (using ON CONFLICT clauses)
- Invalid data
- Transaction failures

If any error occurs, the script will:
1. Print the error message
2. Exit with a non-zero status code
3. Roll back any pending transactions

## Security Considerations

1. The script uses parameterized queries to prevent SQL injection
2. Passwords are stored as hashed values using bcrypt
3. Sensitive data is not logged
4. Database credentials should be stored securely
5. The password hash generator uses a secure cost factor for bcrypt

## Maintenance

To update the initial data:
1. Modify the data structures in the script
2. Run the script again
3. The ON CONFLICT clauses will ensure that existing data is not duplicated

## Troubleshooting

Common issues and solutions:

1. Database Connection Error
   - Verify database credentials
   - Check if database is running
   - Ensure network connectivity

2. Duplicate Entry Errors
   - These are normal and can be ignored
   - The script uses ON CONFLICT clauses to handle duplicates

3. Permission Errors
   - Ensure database user has necessary permissions
   - Check if tables exist and are accessible

4. Transaction Errors
   - Check for foreign key constraints
   - Verify data integrity
   - Check for unique constraints

5. Password Hash Issues
   - Ensure the password hash is properly generated
   - Check that the hash is correctly copied to init_data.go
   - Verify the hash format matches bcrypt requirements 