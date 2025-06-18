# inkube

> Be in the cluster, without being in the cluster. Develop locally with real cluster context.

---

## 🔍 Overview

**inkube** enables developers to run their services locally while mirroring the runtime environment of a Kubernetes pod. This means:

- You get access to the **same environment variables, secrets, and config maps** as the pod.
- Network **traffic to and from the pod is intercepted** and redirected to your local service.
- You can **test, debug, and iterate** locally, with full cluster integration — no redeploys needed.

---

## ✨ Features

- 🧩 **Pod Environment Mirroring**: Clone envs, secrets, ConfigMaps, and volumes from an existing pod.
- 🌐 **Traffic Interception**: Seamlessly redirect traffic between the cluster and your local machine.
- 🔐 **Secrets Mounting**: Automatically mount Kubernetes secrets locally in memory or as files.
- 🧪 **Live Development Mode**: Run and test services locally while communicating with real cluster services.
- 🎯 **Namespace + Context Targeting**: Select specific contexts and namespaces per session.

---

##  Architecture

![Architecture](./arch.png)

> ⚠️ inkube is currently in **development**.



### 📦 Installation

```bash
# install the binary
go install github.com/abdheshnayak/inkube@latest
```

### 🚀 Usage

```bash
# create a new inkube config file
inkube init
```

```bash
# switch to a different namespace and context
inkube switch
```

```bash
# start a live development session
inkube dev
```

```bash
# intercept a running pod
inkube intercept
```

```bash
# leave an intercepted pod
inkube leave
```

```bash
# quit the live development session

exit # when you will exit the shell, the session will be quit

# or
inkube quit
```

---

## 📝 License

Copyright © 2025 [Abhishek Nayak](https://github.com/abdheshnayak).<br />
This project is [MIT](./LICENSE) licensed.
