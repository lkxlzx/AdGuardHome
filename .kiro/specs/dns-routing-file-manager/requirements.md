# Requirements Document

## Introduction

This document specifies the requirements for a centralized DNS routing rule file management module. Currently, DNS routing rule file operations (parsing, loading, saving) are scattered across multiple components (internal/home/dns_routing.go, internal/filtering/filter.go, internal/dnsrouting/parser.go), leading to inconsistent behavior, difficult maintenance, and integration issues. This module will provide a single, unified interface for all DNS routing rule file operations, improving code quality and preventing bugs related to rule file handling.

## Glossary

- **DNS Routing Rule**: A configuration that maps domain patterns to specific upstream DNS server groups
- **Rule File**: A file containing DNS routing rules in various formats (Clash, GFWList, AdGuard)
- **File Manager**: The centralized module responsible for all rule file operations
- **Domain List Rule**: Rules loaded from external URL sources (rule files)
- **Custom Domain Rule**: User-defined rules entered directly through the UI
- **Rule Persistence**: The process of saving rules to disk and loading them on restart
- **Rule Parser**: Component that converts rule files from various formats into internal representation
- **Filter System**: The existing AdGuard Home filtering system that currently handles some rule file operations

## Requirements

### Requirement 1

**User Story:** As a system administrator, I want all DNS routing rule files to be managed by a single module, so that rule file operations are consistent and reliable across the application.

#### Acceptance Criteria

1. WHEN any component needs to save a DNS routing rule file THEN the File Manager SHALL handle the save operation
2. WHEN any component needs to load a DNS routing rule file THEN the File Manager SHALL handle the load operation
3. WHEN any component needs to parse a DNS routing rule file THEN the File Manager SHALL handle the parsing operation
4. WHEN the application starts THEN the File Manager SHALL automatically load all persisted rule files
5. WHEN a rule file is modified THEN the File Manager SHALL ensure the changes are persisted to disk

### Requirement 2

**User Story:** As a developer, I want the File Manager to provide a clean API for rule file operations, so that I can easily integrate rule file functionality without understanding implementation details.

#### Acceptance Criteria

1. WHEN a developer needs to add a new domain list rule THEN the File Manager SHALL provide a single method that handles downloading, parsing, and persisting
2. WHEN a developer needs to update an existing rule THEN the File Manager SHALL provide a single method that handles all update operations
3. WHEN a developer needs to delete a rule THEN the File Manager SHALL provide a single method that handles cleanup and persistence
4. WHEN a developer needs to refresh a rule from its URL THEN the File Manager SHALL provide a single method that handles re-downloading and updating
5. THE File Manager SHALL expose a clear interface that abstracts away file system operations

### Requirement 3

**User Story:** As a system administrator, I want rule files to be stored in a dedicated directory with consistent naming, so that I can easily locate and manage rule files on disk.

#### Acceptance Criteria

1. WHEN a domain list rule is added THEN the File Manager SHALL save the rule file to a dedicated directory (e.g., data/dns_routing_rules/)
2. WHEN saving rule files THEN the File Manager SHALL use consistent naming conventions based on rule ID
3. WHEN the application restarts THEN the File Manager SHALL load all rule files from the dedicated directory
4. WHEN a rule is deleted THEN the File Manager SHALL remove the corresponding file from disk
5. THE File Manager SHALL create the rule directory if it does not exist

### Requirement 4

**User Story:** As a system administrator, I want custom domain rules to be persisted separately from domain list rules, so that user-defined rules are preserved across restarts.

#### Acceptance Criteria

1. WHEN custom domain rules are modified THEN the File Manager SHALL save them to a dedicated file (e.g., data/dns_routing_rules/custom_rules.txt)
2. WHEN the application starts THEN the File Manager SHALL load custom rules from the dedicated file
3. WHEN custom rules are saved THEN the File Manager SHALL use a format that preserves all rule properties (domain, match type, upstream group, enabled state)
4. WHEN the custom rules file is corrupted THEN the File Manager SHALL handle the error gracefully and log the issue
5. THE File Manager SHALL ensure custom rules are saved atomically to prevent corruption

### Requirement 5

**User Story:** As a developer, I want the File Manager to integrate with the existing DNS router, so that rule changes are automatically reflected in the routing behavior.

#### Acceptance Criteria

1. WHEN a rule file is loaded THEN the File Manager SHALL update the DNS router with the parsed rules
2. WHEN a rule is added, updated, or deleted THEN the File Manager SHALL notify the DNS router to reload
3. WHEN multiple rule operations occur in sequence THEN the File Manager SHALL batch router updates to avoid unnecessary reloads
4. WHEN the DNS router is not initialized THEN the File Manager SHALL queue rule updates for later application
5. THE File Manager SHALL provide a callback mechanism for router integration

### Requirement 6

**User Story:** As a system administrator, I want rule file operations to be thread-safe, so that concurrent access does not cause data corruption or race conditions.

#### Acceptance Criteria

1. WHEN multiple goroutines access the File Manager simultaneously THEN the File Manager SHALL ensure thread-safe operations
2. WHEN a rule file is being written THEN the File Manager SHALL prevent concurrent writes to the same file
3. WHEN reading rule files THEN the File Manager SHALL allow concurrent reads
4. WHEN updating rule metadata THEN the File Manager SHALL use appropriate locking mechanisms
5. THE File Manager SHALL use sync primitives (mutexes, RWMutex) to protect shared state

### Requirement 7

**User Story:** As a system administrator, I want rule file parsing to support multiple formats, so that I can use rule lists from various sources.

#### Acceptance Criteria

1. WHEN a rule file is in Clash format THEN the File Manager SHALL parse it correctly
2. WHEN a rule file is in GFWList format THEN the File Manager SHALL parse it correctly
3. WHEN a rule file is in AdGuard format THEN the File Manager SHALL parse it correctly
4. WHEN a rule file format cannot be detected THEN the File Manager SHALL return a clear error message
5. THE File Manager SHALL reuse the existing parser implementation in internal/dnsrouting/parser.go

### Requirement 8

**User Story:** As a system administrator, I want rule file operations to be logged, so that I can troubleshoot issues and monitor rule file activity.

#### Acceptance Criteria

1. WHEN a rule file is saved THEN the File Manager SHALL log the operation with relevant details (rule ID, file path, size)
2. WHEN a rule file is loaded THEN the File Manager SHALL log the operation with relevant details (rule ID, rules count)
3. WHEN a rule file operation fails THEN the File Manager SHALL log the error with context
4. WHEN parsing a rule file THEN the File Manager SHALL log the detected format and number of valid rules
5. THE File Manager SHALL use structured logging with appropriate log levels

### Requirement 9

**User Story:** As a developer, I want the File Manager to handle rule file downloads, so that I don't need to implement HTTP client logic in multiple places.

#### Acceptance Criteria

1. WHEN a domain list rule is added with a URL THEN the File Manager SHALL download the rule file from the URL
2. WHEN downloading a rule file THEN the File Manager SHALL use appropriate timeouts and error handling
3. WHEN a download fails THEN the File Manager SHALL return a clear error message
4. WHEN a rule is refreshed THEN the File Manager SHALL re-download the file from its URL
5. THE File Manager SHALL support HTTP and HTTPS protocols

### Requirement 10

**User Story:** As a system administrator, I want rule file metadata to be persisted, so that rule information (name, URL, update time) is preserved across restarts.

#### Acceptance Criteria

1. WHEN a rule is added THEN the File Manager SHALL save its metadata (ID, name, URL, upstream group, priority, enabled state)
2. WHEN the application restarts THEN the File Manager SHALL load rule metadata from disk
3. WHEN rule metadata is updated THEN the File Manager SHALL persist the changes
4. WHEN a rule file is refreshed THEN the File Manager SHALL update the last_updated timestamp
5. THE File Manager SHALL store metadata in a structured format (JSON or YAML)
