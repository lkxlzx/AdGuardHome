# Implementation Plan

- [x] 1. Create package structure and core types



  - Create `internal/dnsroutingfiles/` package directory
  - Define `DomainListRule` and `CustomRule` structs
  - Define `Metadata` struct for persistence
  - Define `Manager` interface with all required methods
  - Define `Config` struct with configuration options
  - _Requirements: 1.1, 2.5, 10.5_

- [x] 2. Implement metadata persistence


  - [x] 2.1 Implement metadata serialization to JSON


    - Write `saveMetadata()` function to serialize Metadata to JSON
    - Write `loadMetadata()` function to deserialize JSON to Metadata
    - Handle version field for future migrations
    - _Requirements: 10.1, 10.2, 10.5_

  - [x] 2.2 Write property test for metadata round-trip


    - **Property 19: Metadata persistence**
    - **Validates: Requirements 10.1, 10.3**

  - [x] 2.3 Write property test for metadata loading

    - **Property 20: Metadata loading**
    - **Validates: Requirements 10.2**

- [x] 3. Implement file storage operations


  - [x] 3.1 Implement directory initialization

    - Write `ensureDirectory()` function to create rule directory if needed
    - Handle permission errors gracefully
    - _Requirements: 3.5_

  - [x] 3.2 Write property test for directory initialization


    - **Property 7: Directory initialization**
    - **Validates: Requirements 3.5**

  - [x] 3.3 Implement file path generation

    - Write `getRuleFilePath()` function to generate file paths based on rule ID
    - Write `getCustomRulesFilePath()` function for custom rules file
    - Ensure consistent naming conventions
    - _Requirements: 3.1, 3.2_

  - [x] 3.4 Write property test for file storage conventions

    - **Property 5: File storage conventions**
    - **Validates: Requirements 3.1, 3.2**

  - [x] 3.5 Implement atomic file writes

    - Use `aghrenameio.PendingFile` for atomic writes
    - Write `writeFileAtomic()` helper function
    - Handle write failures and cleanup
    - _Requirements: 4.5_

  - [x] 3.6 Write property test for atomic writes

    - **Property 9: Atomic write safety**
    - **Validates: Requirements 4.5**

- [x] 4. Implement custom rules persistence


  - [x] 4.1 Implement custom rules file format

    - Define text format for custom rules (one rule per line)
    - Include all fields: domain, match type, upstream group, enabled
    - Write `formatCustomRule()` and `parseCustomRule()` functions
    - _Requirements: 4.1, 4.3_

  - [x] 4.2 Implement custom rules save/load

    - Write `saveCustomRules()` function to write rules to file
    - Write `loadCustomRules()` function to read rules from file
    - Handle corrupted file errors gracefully
    - _Requirements: 4.1, 4.2, 4.4_

  - [x] 4.3 Write property test for custom rules round-trip

    - **Property 8: Custom rules round-trip**
    - **Validates: Requirements 4.3**

- [x] 5. Implement HTTP download functionality



  - [x] 5.1 Implement rule file downloader


    - Write `downloadRuleFile()` function with timeout support
    - Use provided HTTP client from config
    - Support both HTTP and HTTPS protocols
    - Return clear errors for failures
    - _Requirements: 9.1, 9.2, 9.5_

  - [x] 5.2 Write property test for download functionality

    - **Property 15: Download functionality**
    - **Validates: Requirements 9.1**

  - [x] 5.3 Write property test for protocol support

    - **Property 18: Protocol support**
    - **Validates: Requirements 9.5**

  - [x] 5.4 Write property test for download error handling

    - **Property 16: Download error handling**
    - **Validates: Requirements 9.2**

- [x] 6. Implement File Manager core structure


  - [x] 6.1 Create Manager struct with fields

    - Add fields: config, logger, mutex, domainListRules, customRules
    - Add router update callback field
    - Add HTTP client field
    - _Requirements: 1.1, 6.1_

  - [x] 6.2 Implement Manager constructor

    - Write `NewManager()` function
    - Initialize all fields
    - Call `ensureDirectory()` to create rule directory
    - _Requirements: 1.1, 3.5_

  - [x] 6.3 Implement LoadAll method

    - Load metadata from disk
    - Load all domain list rule files
    - Load custom rules file
    - Populate internal state
    - _Requirements: 1.4, 3.3_

  - [x] 6.4 Write property test for startup loading

    - **Property 2: Startup loading completeness**
    - **Validates: Requirements 1.4, 3.3**

- [x] 7. Implement domain list rule operations



  - [x] 7.1 Implement AddDomainListRule method


    - Validate rule parameters
    - Download rule file from URL
    - Parse rule file using existing parser
    - Save rule file to disk
    - Update metadata
    - Notify router
    - _Requirements: 2.1, 9.1_

  - [x] 7.2 Implement UpdateDomainListRule method

    - Find existing rule by ID
    - Update rule properties
    - Re-download if URL changed
    - Update metadata
    - Notify router
    - _Requirements: 2.2_

  - [x] 7.3 Implement DeleteDomainListRule method

    - Find rule by ID
    - Remove rule file from disk
    - Update metadata
    - Notify router
    - _Requirements: 2.3, 3.4_

  - [x] 7.4 Write property test for file cleanup on deletion

    - **Property 6: File cleanup on deletion**
    - **Validates: Requirements 3.4**

  - [x] 7.5 Implement RefreshDomainListRule method

    - Find rule by ID
    - Re-download rule file from URL
    - Parse and save updated file
    - Update last_updated timestamp
    - Update metadata
    - Notify router
    - _Requirements: 2.4, 9.4, 10.4_

  - [x] 7.6 Write property test for refresh re-downloads

    - **Property 17: Refresh re-downloads**
    - **Validates: Requirements 9.4**

  - [x] 7.7 Write property test for timestamp updates

    - **Property 21: Timestamp updates**
    - **Validates: Requirements 10.4**

  - [x] 7.8 Implement GetDomainListRule and GetAllDomainListRules methods

    - Use read lock for thread safety
    - Return copies of rules to prevent external modification
    - _Requirements: 6.1, 6.3_

- [x] 8. Implement custom rule operations


  - [x] 8.1 Implement AddCustomRule method

    - Validate rule parameters
    - Check for duplicates
    - Add to internal state
    - Save custom rules file
    - Update metadata
    - Notify router
    - _Requirements: 2.1_

  - [x] 8.2 Implement UpdateCustomRule method

    - Find rule by domain and match type
    - Update rule properties
    - Save custom rules file
    - Update metadata
    - Notify router
    - _Requirements: 2.2_

  - [x] 8.3 Implement DeleteCustomRule method

    - Find and remove rule
    - Save custom rules file
    - Update metadata
    - Notify router
    - _Requirements: 2.3_

  - [x] 8.4 Implement GetAllCustomRules method

    - Use read lock for thread safety
    - Return copies of rules
    - _Requirements: 6.1, 6.3_

- [x] 9. Implement router integration

  - [x] 9.1 Implement router update notification

    - Write `notifyRouter()` method
    - Call router update callback if provided
    - Handle callback errors gracefully
    - _Requirements: 5.1, 5.2_

  - [x] 9.2 Write property test for router integration

    - **Property 10: Router integration**
    - **Validates: Requirements 5.1, 5.2**

  - [x] 9.3 Implement batch update mechanism

    - Add update queue and timer
    - Debounce router updates within 100ms window
    - Call router update once after window expires
    - _Requirements: 5.3_

  - [x] 9.4 Write property test for batch router updates

    - **Property 11: Batch router updates**
    - **Validates: Requirements 5.3**

- [x] 10. Implement thread safety

  - [x] 10.1 Add mutex protection to all methods

    - Use write lock for: add, update, delete, refresh
    - Use read lock for: get, list operations
    - Ensure proper lock/unlock in all code paths
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [x] 10.2 Write property test for thread-safe operations

    - **Property 12: Thread-safe operations**
    - **Validates: Requirements 6.1**

  - [x] 10.3 Write property test for concurrent write prevention

    - **Property 13: Concurrent write prevention**
    - **Validates: Requirements 6.2**

  - [x] 10.4 Write property test for concurrent read support

    - **Property 14: Concurrent read support**
    - **Validates: Requirements 6.3**

- [x] 11. Implement comprehensive logging

  - [x] 11.1 Add logging to all operations

    - Log file saves with rule ID, path, size
    - Log file loads with rule ID, rules count
    - Log errors with full context
    - Log parsing results with format and count
    - Use structured logging with appropriate levels
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

- [x] 12. Implement Close method

  - Flush any pending router updates
  - Release resources
  - _Requirements: 1.1_

- [x] 13. Checkpoint - Ensure all tests pass


  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Update HTTP handlers to use File Manager





  - [x] 14.1 Update handleAddDnsRoutingRule


    - Replace direct file operations with File Manager calls
    - Remove filtering system integration
    - _Requirements: 1.1, 2.1_

  - [x] 14.2 Update handleUpdateDnsRoutingRule


    - Replace direct file operations with File Manager calls
    - Remove filtering system integration
    - _Requirements: 1.2, 2.2_

  - [x] 14.3 Update handleDeleteDnsRoutingRule


    - Replace direct file operations with File Manager calls
    - Remove filtering system integration
    - _Requirements: 1.1, 2.3_

  - [x] 14.4 Update handleRefreshDnsRoutingRule


    - Replace direct file operations with File Manager calls
    - Remove filtering system integration
    - _Requirements: 1.3, 2.4_

  - [x] 14.5 Update custom rule handlers


    - Update handleAddCustomDomainRule to use File Manager
    - Update handleUpdateCustomDomainRule to use File Manager
    - Update handleDeleteCustomDomainRule to use File Manager
    - _Requirements: 1.1, 1.2, 1.3_

- [ ] 15. Integrate File Manager with DNS initialization
  - [ ] 15.1 Initialize File Manager in initDNS
    - Create File Manager instance with config
    - Set router update callback
    - Call LoadAll to load persisted rules
    - _Requirements: 1.4, 5.1_

  - [ ] 15.2 Update DNS server to use File Manager rules
    - Modify router initialization to get rules from File Manager
    - Remove direct file loading from filtering system
    - _Requirements: 1.2, 5.1_

- [x] 16. Implement migration from existing implementation


  - [x] 16.1 Detect existing rule files

    - Check for DNS routing rules in `data/filters/` directory
    - Identify rules by DnsRouting flag in config
    - _Requirements: 1.4_

  - [x] 16.2 Migrate existing rules

    - Copy rule files to `data/dns_routing_rules/` directory
    - Generate metadata from existing configuration
    - Preserve all rule IDs and settings
    - _Requirements: 1.4, 10.1_

  - [x] 16.3 Clean up old implementation

    - Remove DNS routing file handling from filtering system
    - Remove old rule files after successful migration
    - Update configuration to remove DnsRouting flags
    - _Requirements: 1.1_

- [ ] 17. Write integration tests
  - Test File Manager with DNS Router integration
  - Test File Manager with HTTP handlers
  - Test end-to-end rule lifecycle
  - Test restart behavior

- [ ] 18. Write property test for API completeness
  - **Property 4: API operation completeness**
  - **Validates: Requirements 2.1, 2.2, 2.3, 2.4**

- [ ] 19. Write property test for modification persistence
  - **Property 3: Modification persistence**
  - **Validates: Requirements 1.5**

- [ ] 20. Write property test for centralized operations
  - **Property 1: Centralized file operations**
  - **Validates: Requirements 1.1, 1.2, 1.3**

- [ ] 21. Final Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
