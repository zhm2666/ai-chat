package filter_test

import (
	"flag"
	"keywords-filter/pkg/filter"
	"testing"
)

var dictFile = flag.String("dict", "../dict.txt", "")
var text = flag.String("text", "", "")

func TestMain(m *testing.M) {
	flag.Parse()
	filter.InitFilter(*dictFile)
	if *text == "" {
		*text = `学习Golang可以按照以下步骤进行：

学习基本语法：了解Golang的基本语法、变量、数据类型、流程控制等。可以通过官方文档或在线教程进行学习。

理解并使用包和模块：Golang具有强大的包管理和模块化机制，学会使用和管理包对于开发是非常重要的。

编写简单程序：通过编写一些简单的程序来练习和熟悉语言特性和语法规则。可以从简单的命令行工具开始，逐渐扩展到更复杂的应用程序。

学习标准库：Golang拥有丰富而强大的标准库，包含各种常用功能。学习如何使用标准库中的各种包，并掌握其使用方法。

实践项目：选择一个小型项目，将所学知识应用于实践中。通过实际项目锻炼自己，在不断实践中提高自己的能力。

阅读优秀代码：阅读别人优秀的Golang代码是一个很好的学习方法。通过阅读源码或开源项目，了解更多最佳实践和设计模式。

参与社区讨论：加入Golang社区，参与讨论、提问和回答问题。与其他开发者交流，分享经验，互相学习。

记住，持续学习和实践是掌握Golang的关键。不断挑战自己，扩展知识面，并尝试解决各种实际问题。`
	}
	m.Run()
}

func BenchmarkValidate(b *testing.B) {
	f := filter.GetFilter()
	for i := 0; i < b.N; i++ {
		f.Validate(*text)
	}
}
func BenchmarkFindAll(b *testing.B) {
	f := filter.GetFilter()
	for i := 0; i < b.N; i++ {
		f.FindAll(*text)
	}
}
