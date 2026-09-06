-- ==============================================================================
-- Webhook Ecosystem MySQL Initialization Script
-- Automatically executes on container first startup in /docker-entrypoint-initdb.d/
-- ==============================================================================

CREATE DATABASE IF NOT EXISTS `webhook_accounts` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `webhook_subscriptions` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `webhook_runner` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `webhook_audit` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Ensure root and application user permissions
GRANT ALL PRIVILEGES ON `webhook_accounts`.* TO 'root'@'%';
GRANT ALL PRIVILEGES ON `webhook_subscriptions`.* TO 'root'@'%';
GRANT ALL PRIVILEGES ON `webhook_runner`.* TO 'root'@'%';
GRANT ALL PRIVILEGES ON `webhook_audit`.* TO 'root'@'%';

FLUSH PRIVILEGES;
