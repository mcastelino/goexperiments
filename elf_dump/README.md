# ELF Parsing leveraging templates

By default

elf_dump <filename> just dumps the Program Headers and the Section Headers as well the symbol table

However what makes this interesting the ability to use templates to do interesting stuff

An few examples

You can filter the symbol table to extract information in different ways

```
elf_dump -f '{{table (filterHasPrefix  (getSymbols .) "Name" "sys")}}' vmlinux
```

```
elf_dump -f '{{table (cols (filterHasPrefix  (getSymbols .) "Name" "sys") "Name" "Value")}}' vmlinux
```
