# Haniel

This Project name form [なつみ（七罪）的天使 贗造魔女（Haniel）](https://zh.moegirl.org/zh-hans/镜野七罪)

> あぁ、素直に笑えたら——もう、そんな事、赦されるはずもないだろ！

> Let's learn meowing together: meow~ meow~ meow~ meow~ meow~ meow~ meow~

Socket(Server/Client) IO simulation

[The Documentation](fixtures.yaml)


* golang >= 1.13


```sh
./haniel -h
Usage of ./haniel:
-c string
the fixtures config (default "fixtures.yaml")
-h    Show help
-l string
As socket Server Address default enable (default "localhost:1234")
-log string
the running log path (default "haniel.log")
-p string
As socket Client Address default disable '-p || -l'
```

```sh
./haniel -c fixtures.yaml -l localhost:8900


# As tcp client
nc -l 1234
./haniel -c fixtures.yaml -p localhost:1234
```

