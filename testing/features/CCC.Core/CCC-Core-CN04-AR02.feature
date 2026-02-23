# @PerService @CCC.Core @CCC.Core.CN04 @tlp-amber @tlp-red
# Feature: CCC.Core.CN04.AR02 - Log Data Modification Attempts
#   As a security administrator
#   I want to ensure all data modification attempts are logged
#   So that data changes are auditable

#   Background:
#     Given a cloud api for "{Provider}" in "api"

#   @Policy @object-storage @vpc
#   Scenario: Data modification logging compliance
#     When I attempt policy check "data-write-logging" for control "CCC.Core.CN04" assessment requirement "AR02" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
#     Then "{result}" is true

#   @Behavioural @object-storage
#   Scenario: Verify data modifications are logged with identity and timestamp
#     Given I call "{api}" with "GetServiceAPI" using argument "object-storage"
#     And I refer to "{result}" as "storage"
#     When I call "{storage}" with "CreateObject" using arguments "{ResourceName}", "test-logging-object.txt", and "test data for logging verification"
#     Then "{result}" is not an error
#     And I refer to "{result}" as "createResult"
#     And I attach "{createResult}" to the test output as "Object Create Result"
#     When I call "{storage}" with "DeleteObject" using arguments "{ResourceName}" and "test-logging-object.txt"
#     Then "{result}" is not an error
#     And I refer to "{result}" as "deleteResult"
#     And I attach "{deleteResult}" to the test output as "Object Delete Result"
#     When I call "{storage}" with "QueryDataWriteLogs" using arguments "{ResourceName}" and "60"
#     Then "{result}" is not an error
#     And I refer to "{result}" as "dataLogs"
#     And I attach "{dataLogs}" to the test output as "Data Write Logs"
#     And "{dataLogs}" is an array of objects with at least the following contents
#       | Identity  | Action              |
#       | .+        | .*(Create\|Put).*   |
