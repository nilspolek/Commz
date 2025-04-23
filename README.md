# Commz

## Setup:
#### Clone the repository:
```
git clone --recurse-submodules ssh://git@team6-managing.mni.thm.de:222/Commz/infrastructure.git
```
#### Update submodules:
```
cd infrastructure
git submodule update --remote --merge
```

#### Run the infrastructure:
```
cd infrastructure
docker compose build
docker compose up -d
```

#### Access Commz
[http://localhost](http://localhost)
