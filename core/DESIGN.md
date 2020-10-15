# Design

## Eval

```python
def eval(env: Env, form: Any) -> Any:
    expr = analyze(env, form)
    return expr.eval()
```

## Analyze

```python
def analyze(env: Env, form: Any) -> Expr:
    if isinstance(form, [int, str, bool, char, float, keyword]):
        return LiteralExpr{Val: form}

```
