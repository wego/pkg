# pkg
Wego common Go packages

This repository contains multiple modules. Each module can be developed, tested & released independently.

E.g.

`git tag common/v0.2.3`

`git tag errors/v0.1.2`

# Setup local environment

```
git config core.hookspath .githooks
sudo chmod +x .githooks/*
```

# Tag a new version

After merge your PR into `main` branch, run this

```
./auto_version
```

If it's your first time, you might need to run this first

```
chmod +x auto_version
```

---

References:
- https://golang.org/doc/modules/managing-source#multiple-module-source
- https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories