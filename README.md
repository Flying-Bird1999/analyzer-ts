# analyzer-ts

识别 .ts/.tsx 文件中的类型，找到其依赖关系，最终给出结果

```typescript
type Name = {
    age: number;
    name: string;
    class: Class;
}

type School = {
    area: string;
}

interface Class {
    teacher: string;
    count: number;
}

```
====>

```typescript
type Name = {
    age: number;
    name: string;
    class: Class;
}

interface Class {
    teacher: string;
    count: number;
}

```