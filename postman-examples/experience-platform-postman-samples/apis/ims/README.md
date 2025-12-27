# Identity Management Service APIs

This set of Postman Collections contains commonly used interactions with [Adobe Identity Management Services (IMS) APIs](https://www.adobe.io/authentication/auth-methods.html#!AdobeDocs/adobeio-auth/master/AuthenticationOverview/AuthenticationGuide.md)

These Postman Collections have been set up to share the Postman Variable collections generated via [Adobe Developer Console's Integrations](https://console.adobe.io/integrations) -> Export Details to Postman, which generates a Postman Environment file complete with your Integrations values.

After Exporting the Postman Environment from Adobe I/O Console, ensure you add your Private Key (`-----BEGIN PRIVATE KEY----- ... -----END PRIVATE KEY-----`) to BOTH the Initial and Current values for `PRIVATE_KEY`.


## Obtaining an Access Token to access the Adobe Experience Platform APIs for Non-production Use

The Identity Management Service Postman collection contains a single option for generating an Access Token. This method has ben built for develop convenience and provides a quick way for developers to obtain an Access Token for non-production use. The reasons these are for non-production use are described below:

- Local signing leverages a 3rd party JSR Assign Crypto library to be loaded and locally sign the JWT Token using the provided Private Key
- Using this method the Private Key never leaves the local machine, however 3rd party JavaScript is loaded into the Postman context.


