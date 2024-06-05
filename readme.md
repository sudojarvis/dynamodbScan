# README: Setting Up DynamoDB Permissions and Cross-Account Access

## Instructions


1. **Attach an Inline Policy to the User**
    - On the **Permissions** page, click **Add inline policy**.
    - Select the **JSON** tab.
    - Paste the policy below, replacing `{target_account}` with the target account ID:
        ```json
        {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Sid": "Statement1",
                    "Effect": "Allow",
                    "Action": [
                        "sts:AssumeRole",
                        "iam:PassRole"
                    ],
                    "Resource": [
                        "arn:aws:iam::{target_account}:role/{role_name}"
                    ]
                }
            ]
        }
        ```
   

2. **Create a Role in the Target Account**
    - Sign in to the AWS Management Console with target account credentials.
    - Go to **IAM** > **Roles** > **Create role**.
    - Select **Another AWS account**.
    - Enter the source account ID.
    - Click **Next: Permissions**.

3. **Attach a Trust Policy to the Role**
    - On the **Permissions** page, create an inline policy:
        - Click **Create policy**, select the **JSON** tab, and paste the policy below, replacing `{source_id}` with the source account ID:
            ```json
            {
                "Version": "2012-10-17",
                "Statement": [
                    {
                        "Sid": "Statement1",
                        "Effect": "Allow",
                        "Principal": {
                            "AWS": "arn:aws:iam::{source_id}:user/{source_user}"
                        },
                        "Action": "sts:AssumeRole"
                    }
                ]
            }
            ```
        - Click **Review Policy**.
        - Name the policy `TrustPolicy`.
        - Click **Create policy**.
    - Attach `AmazonEC2FullAccess` or the newly created `TrustPolicy`.
   

   ### Resources
    - [Cross-Account Access](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html)
    - [ACCESS RESOURCES USING IAM](https://repost.aws/knowledge-center/cross-account-access-iam)
    - [STACK OVERFLOW](https://stackoverflow.com/questions/73206798/launch-ec2-instance-using-iam-role-on-multiple-aws-accounts)