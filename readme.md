### Cross-Account Setup

#### For Customer Side:

1. **Create a new role with the following trust relationship:**

   Replace `{source_id}` with the AWS account ID of the source account and `{source_user}` with the username of the source user.

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

2. **Attach the following policy to the role:**

   ```json
   {
       "Version": "2012-10-17",
       "Statement": [
           {
               "Effect": "Allow",
               "Action": [
                   "ec2:DescribeInstances",
                   "ec2:StartInstances",
                   "ec2:StopInstances",
                   "ec2:RebootInstances",
                   "ec2:TerminateInstances",
                   "ec2:DescribeVolumes",
                   "ec2:DescribeTags",
                   "ec2:DescribeInstanceStatus",
                   "ec2:DescribeVolumeStatus",
                   "ec2-instance-connect:SendSSHPublicKey",
                   "ec2:DescribeSecurityGroups",
                   "ec2:DescribeRouteTables",
                   "ec2:DescribeSubnets",
                   "ec2:DescribeVpcs",
                   "ec2:RunInstances",
                   "ec2:DescribeImages",
                   "ec2:CreateTags",
                   "ec2:DescribeInstanceTypes"
               ],
               "Resource": "*"
           },
           {
               "Effect": "Allow",
               "Action": [
                   "dynamodb:CreateBackup",
                   "dynamodb:ListBackups",
                   "dynamodb:DescribeBackup",
                   "dynamodb:DeleteBackup",
                   "dynamodb:RestoreTableFromBackup",
                   "dynamodb:DescribeTable",
                   "dynamodb:CreateTable",
                   "dynamodb:ListTables",
                   "dynamodb:Scan",
                   "dynamodb:UpdateTable",
                   "dynamodb:DeleteTable",
                   "dynamodb:BatchWriteItem",
                   "dynamodb:PutItem",
                   "dynamodb:DeleteItem",
                   "dynamodb:BatchGetItem",
                   "dynamodb:GetItem",
                   "dynamodb:Query"
               ],
               "Resource": "*"
           }
       ]
   }
   ```

3. **Attach the ARN of the role to the user policy in the source account to switch roles.**

#### For Service Side:

1. **Attach the ARN of the role created in the customer account:**

   Replace `{customer_id}` with the AWS account ID of the customer account and `{role_name}` with the name of the role created.

   ```json
   {
       "Version": "2012-10-17",
       "Statement": [
           {
               "Effect": "Allow",
               "Action": "sts:AssumeRole",
               "Resource": "arn:aws:iam::{customer_id}:role/{role_name}"
           }
       ]
   }
   ```