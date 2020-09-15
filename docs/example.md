## Prerequisite

1. Create a GKE cluster with workload identity enabled

```bash
gcloud beta container clusters create $CLUSTER_NAME \
  --addons=HttpLoadBalancing,HorizontalPodAutoscaling,CloudRun \
  --machine-type=n1-standard-4 \
  --enable-autoscaling --min-nodes=3 --max-nodes=10 \
  --no-issue-client-certificate --num-nodes=3 --image-type=cos \
  --enable-stackdriver-kubernetes \
  --scopes=cloud-platform,logging-write,monitoring-write,pubsub \
  --zone $CLUSTER_LOCATION \
  --release-channel=rapid
  --workload-pool=$PROJECT.svc.id.goog
```

2. Follow the [instructions](https://cloud.google.com/config-connector/docs/how-to/install-upgrade-uninstall) to install and setup KCC for a namespace `foo`.

3. Create a KSA 

```bash
kubectl create sa foo-runner -n foo
kubectl annotate sa foo-runner -n foo iam.gke.io/gcp-service-account=foo-runner@$PROJECT.iam.gserviceaccount.com
```

3. Prepare a GSA `foo-runner@$PROJECT.iam.gserviceaccount.com` and grant it Cloud SQL and Cloud Storage permissions. And allow the KSA to use it.

```bash
gcloud iam service-accounts add-iam-policy-binding \
--role roles/iam.workloadIdentityUser \
--member serviceAccount:$PROJECT_ID.svc.id.goog[foo/foo-runner] \
foo-runner@$PROJECT_ID.iam.gserviceaccount.com
```

## Install the gcp-binding prototype

```bash
# In repo root
ko apply -f ./config
```

Wait until the controller and webhook are running.

```bash
kubectl get all -n gcp-backing
```

## Provision (with KCC) Cloud SQL databases and Storage buckets

### Databases

Update [databases.yaml](../sample/databases.yaml) with proper names and `kubectl apply -f`. It should provision two Cloud SQL instances/databases/users.

(Pretty painful) Follow the [steps](https://cloud.google.com/sql/docs/mysql/connect-admin-proxy) to manually connect the databases and create a `pet` table.

```sql
-- choose the DB
use {DB_NAME};
-- create a pet table
CREATE TABLE pet (name VARCHAR(20), owner VARCHAR(20), species VARCHAR(20), sex CHAR(1), birth DATE, death DATE);
-- insert some data
INSERT INTO pet (name, owner, species, sex, birth) values ("cake", "chen", "dog", "M", "2020-01-01");
```

### Buckets

Update [buckets.yaml](../sample/buckets.yaml) with proper names and `kubectl apply -f`. It should provision two buckets.

Update pets images to the bucket with `{pet_name}.jpg` as the object name. E.g. `cake.jpg`. And do that for each pet.

## Create bindings

Update [bindings-tmpl.yaml](../sample/bindings-tmpl.yaml) with proper names and `kubectl apply -f`.

## Create sample app

```bash
cd $GOPATH/src/github.com
mkdir yolocs
git clone git@github.com:yolocs/foo-app.git && cd foo-app
ko apply -f ./config/plain
```

The database parameters and bucket name should be injected the application automatically.