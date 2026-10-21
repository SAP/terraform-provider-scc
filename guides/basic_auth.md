### <u> Basic Authentication </u>

> **Breaking change notice**
>
> As of the latest provider version, `http://` URLs are **not permitted** when Basic Auth credentials are configured.
> Basic Auth sends credentials as a base64-encoded header — transmitting them over a plain HTTP connection exposes them in clear text on the network.
>
> If you were using `instance_url = "http://..."` with `username`/`password`, you will now see:
> ```
> Error: Insecure instance_url
> Basic Auth credentials must not be sent over a plain HTTP connection.
> Use an https:// URL or switch to certificate-based authentication.
> ```
>
> **To fix:** change `instance_url` to `https://`, or switch to [certificate-based authentication](./cert_auth.md).
> If the Cloud Connector uses a self-signed certificate, set `skip_ssl_validation = true` temporarily while you obtain the CA certificate — do not leave this enabled in production.

You would require a valid **username** and **password** of an [Administrator](https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/cloud-connector-initial-configuration#loiodb9170a7d97610148537d5a84bf79ba2__log_in) to get authenticated.
 
There are multiple ways to configure your credentials.

1. You can configure them as part of the provider configuration as shown below:

    ```terraform
    provider "scc" {
        instance_url = <your_instance_url>
        username = <your_username>
        password = <your_password>
    }
    ```

2. You can export them as environment variables as shown below:

    #### Windows 

    If you use Windows CMD, do the export via the following commands:

    ```Shell
    set SCC_USERNAME=<your_username>
    set SCC_PASSWORD=<your_password>
    ```

    If you use Powershell, do the export via the following commands:

    ```Shell
    $Env:SCC_USERNAME = '<your_username>'
    $Env:SCC_PASSWORD = '<your_password>'
    ```

    #### Mac

    For Mac OS export the environment variables via:

    ```Shell
    export SCC_USERNAME=<your_username>
    export SCC_PASSWORD=<your_password>
    ```

    #### Linux

    For Linux export the environment variables via:

    ```Shell
    export SCC_USERNAME=<your_username>
    export SCC_PASSWORD=<your_password>
    ```