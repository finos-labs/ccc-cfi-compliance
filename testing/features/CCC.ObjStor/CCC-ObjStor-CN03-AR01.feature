@CCC.ObjStor @tlp-amber @tlp-red @CCC.ObjStor.CN03.AR01
Feature: CCC.ObjStor.CN03.AR01 - Bucket Soft Delete and Recovery
  When an object storage bucket deletion is attempted,
  the bucket MUST be fully recoverable for a set time-frame after deletion is requested.
  
  This ensures protection against accidental bucket deletion.

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"

  Scenario: Service supports bucket soft delete and recovery
    When I call "{storage}" with "CreateBucket" with parameter "ccc-test-soft-delete"
    Then "{result}" is not an error
    And I refer to "{result}" as "testBucket"
    And I attach "{result}" to the test output as "created-bucket.json"
    When I call "{storage}" with "DeleteBucket" with parameter "ccc-test-soft-delete"
    Then "{result}" is not an error
    When I call "{storage}" with "ListDeletedBuckets"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "deleted-buckets.json"
    And "{result}" should have length greater than "0"
    When I call "{storage}" with "RestoreBucket" with parameter "ccc-test-soft-delete"
    Then "{result}" is not an error
    When I call "{storage}" with "ListBuckets"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "restored-buckets.json"
    When I call "{storage}" with "DeleteBucket" with parameter "ccc-test-soft-delete"
    Then "{result}" is not an error
