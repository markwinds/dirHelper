# dirHelper
命令行目录收藏夹 方便切换目录

## 使用方式

1. 运行一次dirHelper -i ddd初始化dirHelper功能, 该条命令将在你的~/.bashrc ~/.zshrc文件中注入dirHelper需要的信息。当然你可以将ddd替换为任何你更喜欢的名字,后续我们都会使用这个名字跳转目录
2. 重新打开一个会话窗口，或者source ~/.bashrc刷新
3. 可以按以下命令使用，ddd就是第一步-i指定的参数
```shell
# 查看收藏夹 你也可以直接查看~/.dirHelper/dirHelper.json
ddd

# 跳转到指定目录
ddd <目录键值>

# 向收藏夹添加目录
ddd -a /home/mark
ddd -a ./

# 删除目录
ddd -r <目录键值>

# 修改目录键值
ddd -c <旧目录键值>,<新目录键值>
```
