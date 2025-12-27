## Adobe Experience Platform Postman Samples

The purpose of this repository is to collect Postman samples that show how to interact with Adobe Experience Platform. Each of the samples will use a common Postman Environment file in order to make it easier to switch between separate collections.

## Install and Setup your Postman App

### Generate and Import the Postman Environment

1. [Install Postman](https://www.getpostman.com/apps)
2. Launch Postman
3. Login to [Adobe I/O Console](https://console.adobe.io/), select _Integrations_, and select the Adobe Experience Platform API Integration you wish to interact with using Postman.
4. Click the __Export Details to Postman__ button to have a Postman Environment file with the Integration's details pre-populated.
5. Select the downloaded __Platform API.postman_environment.json__ file to import the environment.
6. Now click on the newly imported `Platform API`.
7. All of the values with the exception of `PRIVATE_KEY` should be pre-populated.
8. Copy the contents of your `private.key` and use it as the value for `PRIVATE_KEY` in BOTH the Initial and Current values fields.

   **For MacOS & Linux platform**

   From the same terminal you ran `openssl`, execute the following command:

   ```shell
   pbcopy < private.key
   ```

   **For Windows Platform**

   From the same terminal you ran `openssl`, execute the following command:

   ```shell
   notepad private.key
   ```

   Copy the entire key to the keyboard, including the `-----BEGIN PRIVATE KEY-----` and `-----END PRIVATE KEY-----` lines.

1. Click `Update` and close the `Manage Environments` dialog.
1. Now make sure you select the `Platform API` from the environments drop down at the top right of Postman.

### Import the Collection

To import the full Adobe Experience Platform Postman Collection do the following:

1. Download the [`Adobe Experience Platform.postman_colletion.json`](/apis/Adobe%20Experience%20Platform.postman_collection.json) file from this Github repository
1. In Postman, select `Import` in the upper left of the interface
1. Select `Import File` and reference the `Adobe Experience Platform.postman.collection.json` you downloaded in step 1
1. You should then see a collection appear in the `Postman Collections` named `Adobe Experience Platform`


## Reference

See also the [Medium post on the Adobe Techblog](https://medium.com/adobetech/using-postman-for-jwt-authentication-on-adobe-i-o-7573428ffe7f) around the usage of Postman with Adobe I/O.

### Contributing

Contributions are welcomed! Read the [Contributing Guide](CONTRIBUTING.md) for more information.

### Licensing

This project is licensed under the MIT. See [LICENSE](LICENSE) for more information.
