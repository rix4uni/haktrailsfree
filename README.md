## haktrailsfree

Free securitytrails apikey only gives 2k subdomains, you can get 10k subdomains using your cookies, Collect cookie in `https://securitytrails.com/list/apex_domain/google.com`

## Installation
```
go install github.com/rix4uni/haktrailsfree@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/haktrailsfree/releases/download/v0.0.1/haktrailsfree-linux-amd64-0.0.1.tgz
tar -xvzf haktrailsfree-linux-amd64-0.0.1.tgz
rm -rf haktrailsfree-linux-amd64-0.0.1.tgz
mv haktrailsfree ~/go/bin/haktrailsfree
```
Or download [binary release](https://github.com/rix4uni/haktrailsfree/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/haktrailsfree.git
cd haktrailsfree; go install
```

## Usage
```
Usage of haktrailsfree:
  -H string
        User-Agent header (default "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
  -cf string
        Path to cookie file (default: ~/.config/haktrailsfree/cookie.txt or ./cookie.txt)
  -delay int
        Delay between requests in seconds (not recommended to lower delay) (default 3)
  -silent
        Silent mode.
  -version
        Print the version of the tool and exit.
```

## Output Examples

Single URL:
```
echo "google.com" | haktrailsfree
```

Multiple URLs:
```
cat subs.txt | haktrailsfree
```