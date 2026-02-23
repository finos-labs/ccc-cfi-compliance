# @PerService @CCC.Core @CCC.Core.CN04 @tlp-red
# Feature: CCC.Core.CN04.AR03 - Log Data Read Attempts
#   As a security administrator
#   I want to ensure all data read attempts are logged
#   So that data access is fully auditable

#   Background:
#     Given a cloud api for "{Provider}" in "api"

#   @Policy @object-storage @vpc
#   Scenario: Data read logging compliance
#     When I attempt policy check "data-read-logging" for control "CCC.Core.CN04" assessment requirement "AR03" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
#     Then "{result}" is true

#   @Behavioural @object-storage
#   Scenario: Verify data read operations are logged with identity and timestamp
#     Given I call "{api}" with "GetServiceAPI" using argument "object-storage"
#     And I refer to "{result}" as "storage"
#     When I call "{storage}" with "CreateObject" using arguments "{ResourceName}", "test-read-logging-object.txt", and "test data for read logging verification"
#     Then "{result}" is not an error
#     And I refer to "{result}" as "createResult"
#     When I call "{storage}" with "ReadObject" using arguments "{ResourceName}" and "test-read-logging-object.txt"
#     Then "{result}" is not an error
#     And I refer to "{result}" as "readResult"
#     And I attach "{readResult}" to the test output as "Object Read Result"
#     When I call "{storage}" with "DeleteObject" using arguments "{ResourceName}" and "test-read-logging-object.txt"
#     Then "{result}" is not an error
#     When I call "{storage}" with "QueryDataReadLogs" using arguments "{ResourceName}" and "60"
#     Then "{result}" is not an error
#     And I refer to "{result}" as "readLogs"
#     And I attach "{readLogs}" to the test output as "Data Read Logs"
#     And "{readLogs}" is an array of objects with at least the following contents
#       | Identity  | Action           |
#       | .+        | .*(Get\|Read).*  |
