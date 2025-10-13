# Install Steampipe

```
brew install turbot/tap/steampipe
steampipe plugin install aws
steampipe plugin install azure
steampipe plugin install gcp
steampipe service start
```

GCP Account Setup

````
gcloud iam service-accounts create steampipe-reader --project=nodal-time-474015-p5
gcloud projects add-iam-policy-binding nodal-time-474015-p5 \
    --member="serviceAccount:steampipe-reader@nodal-time-474015-p5.iam.gserviceaccount.com" \
    --role="roles/viewer"
 gcloud iam service-accounts keys create ~/steampipe-key.json \
    --iam-account=steampipe-reader@nodal-time-474015-p5.iam.gserviceaccount.com

gcloud projects add-iam-policy-binding nodal-time-474015-p5 --member="serviceAccount:steampipe-reader@nodal-time-474015-p5.iam.gserviceaccount.com" --role="roles/cloudasset.viewer"
    ```
````
