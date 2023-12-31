# ebnf-based-generator

整体效果类似于： 

- https://www.fuzzingbook.org/html/Grammars.html  fuzzing book 的syntactic fuzzing和semantic fuzzing章节

- https://github.com/renatahodovan/grammarinator

- https://github.com/googleprojectzero/domato


## 需求功能点

1. 设计一个上下文无关文法的中间结构表示
  - 该结构附带转换功能：文本语法定义->中间表示，和 中间表示->文本语法定义。
  - 该结构应该能够迅速地转换成一个图结构，或本身就是一个图结构
2. 能够根据已有的token，例如`SELECT * FROM`或空token，首先parse到对应的状态机，然后进行生成操作：
  - 选择non-terminal: 开放一个接口，用于定义选择的算法（例如基于随机，或基于某种概率分布选择）。
  - 生成terminal: 开放一个接口，用于定义生成terminal的算法（例如基于随机、或基于sample、或基于概率）。
  - 能够拦截指定的规则，对于每个状态转移边，具备pre_operation和post_operation的hook。即pre_operation是在转移之前被调用，post_operation是在转移之后被调用。
  - 在选择和生成的过程中，记录语法的覆盖率。例如维护一个map，记录某个状态经历了多少次，某条边走了多少次。
  - 能够打印上述选择、生成过程中走过的路径，以及最后生成结果的语法树。
3. 开放接口，能够指定constraint，并且尽可能地向constraint靠拢。例如指定一定要走某个语法树的路径、一定要经过某个状态、不经过某个状态、生成某个终止符等等。
  - solve constraint过程可能需要用到一些solver。
4. 生成和选择的过程设置为有两种模式，一种是单纯生成（即上面提到的），另一种是reduce。reduce即只生成给定语法路径下的部分路径（随机删减语法图的部分路径），这是为了在找到触发bug的语句时尽可能地自动删减与bug无关的语法，从而减少reduce的时间。reduce也可以被认为是一种constraint
 
### 分析

- 第一点是基础功能，维护一个中间表示，能够解析并转换常见的上下文无关文法。其中，图结构用于后续关联分析什么样的语法结构更容易触发bug，或进行一些优化。
- 第二点是核心功能，基于上下文无关文法的中间表示，给定一个起始符，能够基于默认的算法，或自定义的算法来进行选择和生成逻辑。并且能够在此过程中进行拦截或修改或注入某些自定义功能。同时记录相关运行信息（目前能记的全部记下来），例如覆盖率，经过的语法路径。覆盖率和语法路径可用于后续的分析，例如语法路径可以用于在复现bug的时候指定，从而更快地触发类似的bug。
- 第三点也是核心功能，用于给定约束，然后尽可能地往约束靠拢，形成一个指向性地生成功能。
- 第四点是第三点的扩充，算是一种特殊的constraint。

在上述搭建完毕之后，一个核心逻辑可以被描述为：

1. 随机生成基础的语句（或指定要生成的某些语法），要求覆盖尽可能多的语法
2. 若触发bug，则保存对应信息。并可以对语法进行筛减，保留能同样触发bug的最小的语法树。标记该语法树是一颗重点语法树。
3. 对重点语法树进行一些变异操作，根据局部性原理，可能会有更多的bug被发现。
4. 如果反复触发了非常多同类型的bug，则减少这部分的探索力度，去探索别的语法树。
5. reduce bug并上报开发者

其中，上述的核心逻辑中会有一些参数，例如如何筛减语法树，如何变异，如何选择变异算子等等，这部分后续可以替换成一个强化学习模型，让他实时生成一个概率分布来控制生成器的行为，自动地随着时间的增长来提高适应度。




### cypher enbf示例

https://opencypher.org/resources/

### 目标

基于EBNF语法定义，生成符合定义的语句（代码），并要求尽可能地暴露可控制的参数。

## 示例

以生成SQL为例，下方是用于演示的表格以及数据

```c
mysql> SELECT * FROM Websites;
+----+--------------+---------------------------+-------+---------+
| id | name         | url                       | alexa | country |
+----+--------------+---------------------------+-------+---------+
| 1  | Google       | https://www.google.cm/    | 1     | USA     |
| 2  | 淘宝          | https://www.taobao.com/   | 13    | CN      |
| 3  | 菜鸟教程      | http://www.runoob.com/    | 4689  | CN      |
| 4  | 微博          | http://weibo.com/         | 20    | CN      |
| 5  | Facebook     | https://www.facebook.com/ | 3     | USA     |
+----+--------------+---------------------------+-------+---------+
```

1. 设置起始符号：SELECT

当前状态是SELECT集，下一个TOKEN可以是 id, name, url, alexa, country, * 中的任意一个。

要求选择的下一个TOKEN可以随机，也可以自行指定选择算法（包括自定义算法，或者硬编码返回数据）

1. 假设当前的生成Query已经是：SELECT url FROM Websites WHERE

则要求WHERE的下一个token一定是之前出现过的（在sql中是这么要求的，在pl里可能是要求后续使用的变量一定是前面定义的，且类型得相同）。

例如生成 `WHERE url="https://www.google.cm"` 则是合法的。其中， `"https://www.google.cm"`

可以是任意从数据库里sample出来的值。

第二点描述的是比较具体的针对SQL的情况，在实际实现的更加通用的结构里，这部分应该暴露出一个接口，让用户自定义自己的constraint逻辑，以及自定义生成的逻辑。


