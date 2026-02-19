# CCC.Core.CN03 - Implement Multi-factor Authentication (MFA) for Access

| Field                 | Value                                                                                                                           |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Identity and Access Management                                                                                                  |
| **Control ID**        | CCC.Core.CN03                                                                                                                   |
| **Control Title**     | Implement Multi-factor Authentication (MFA) for Access                                                                          |
| **Control Objective** | Ensure that all sensitive activities require two or more identity factors during authentication to prevent unauthorized access. |

## Assessment Requirements

| Assessment ID      | Requirement Text                                                                                                                                                                                                    |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CCC.Core.CN03.AR01 | When an entity attempts to modify the service through a user interface, the authentication process MUST require multiple identifying factors for authentication.                                                    |
| CCC.Core.CN03.AR02 | When an entity attempts to modify the service through an API endpoint, the authentication process MUST require a credential such as an API key or token AND originate from within the trust perimeter.              |
| CCC.Core.CN03.AR03 | When an entity attempts to view information on the service through a user interface, the authentication process MUST require multiple identifying factors from the user.                                            |
| CCC.Core.CN03.AR04 | When an entity attempts to view information on the service through an API endpoint, the authentication process MUST require a credential such as an API key or token AND originate from within the trust perimeter. |
