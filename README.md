# Waydroid Icon Settings

This is an app to easily toggle hide/show desktop icon of the installed application in waydroid on your linux.

## Stack Used

Golang for the main backend doing tasks such as listing directory, reading files, and give the information back to the client.
For the client, I'm still unsure.

## Running

```bash
# hide
go run .

# show
go run . --hide=false
```

Depending on your system, you might want to remove/flush the cache in your desktop environment, here is example on how can do it in KDE Plasma 6.

```bash
# using cli
kbuildsycoca6 --noincremental

# or directly delete the cache file
rm ~/.cache/ksycoca*
```