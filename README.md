# email
Fetch and store email
修改 config.json 
程序会将对应的 box 的邮件拉取下来并存放, 文件命名为 {seq}.eml
每次执行会从文件夹中最后一个 seq.eml 开始, 即最后一封邮件会重新拉取一次, 防止意外导致的数据缺失