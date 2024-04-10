<!-- toc number-sections -->

This article is part of the *Beginner's Guide to Elvish* series:

-   [Your first Elvish commands](first-commands.html)

-   [Arguments and outputs](arguments-and-outputs.html)

-   [Variables and loops](variables-and-loops.html)

-   [Pipelines and IO](pipelines-and-io.html)

-   **Value types**

-   [Organizing and reusing code](organizing-and-reusing-code.html)

# Maps

We have learned how you can use `curl` to request a URL and `from-json` to
convert JSON-encoded bytes to Elvish values. Combining these two features allows
us to import data from online JSON APIs: for example, let's use the API from
<https://myip.com> to query our IP address and country:

```elvish-transcript Terminal - elvish
~> curl -s https://api.myip.com
{"ip":"10.0.0.31","country":"Elvendom","cc":"EL"}
~> curl -s https://api.myip.com | from-json
▶ [&cc=EL &country=Elvendom &ip=10.0.0.31]
```

The result is surrounded by `[` and `]`, just like lists; but rather than a
list, it's actually a new type of data structure called **maps**.

Lists consist of elements. Maps, on the other hand, consists of *pairs* of
**keys** and **values**. In Elvish, maps are written like `&key=value`, and
putting writing inside `[` and `]` make the overall data structure a map. (The
[general concept of maps](https://en.wikipedia.org/wiki/Associative_array) is
present in many other languages. In this case, our map is actually converted
from a JSON "object", which you can also think of as a map.)

The key-value structure is useful because if there's a key you know, you can
find the corresponding value by **indexing** the map:

```elvish-transcript Terminal - elvish
~> curl -s https://api.myip.com | put (from-json)[country]
▶ Elvendom
```

The `[country]` used for indexing uses the same `[` and `]` for writing lists
and maps, but has a different meaning in this context. You can even use them
together:

```elvish-transcript Terminal - elvish
~> put [&country=Elvendom][country]
▶ Elvendom
```

Here, the first pair of `[]` delimits a map, and the second pair delimits the
index.

To examine the map without having to make the same request every time, we can
save it in a variable:

```elvish-transcript Terminal - elvish
~> curl -s https://api.myip.com | var info = (from-json)
~> put $info[country]
▶ Elvendom
~> put $info[cc]
▶ EL
```

# List indexing

We can also use the indexing syntax to retrieve an element of a list by its
position:

```elvish-transcript Terminal - elvish
~> var triumvirate = [Julius Crassus Pompey]
~> echo $triumvirate[0]' is the most powerful'
Julius is the most powerful
```

The index [starts at zero](https://en.wikipedia.org/wiki/Zero-based_numbering),
like in many other programming languages.

Instead of just one element, lists also allow you to retrieve a **slice** of
elements. There are two variants: *i*..*j* starts from *i* and doesn't include
*j*, while *i*..=*j* includes *j*:

```elvish-transcript Terminal - elvish
~> put $triumvirate[0..2]
▶ [Julius Crassus]
~> put $triumvirate[0..=2]
▶ [Julius Crassus Pompey]
```

As we can see, the result is another list. This is true even if the result is
one or even zero element:

```elvish-transcript Terminal - elvish
~> put $triumvirate[0..1]
▶ [Julius]
~> put $triumvirate[0..0]
▶ []
```

# Nesting data structures

Lists and maps in Elvish can be arbitrarily *nested*. You can have a list of
lists, for example to represent tabular data:

```elvish-transcript Terminal - elvish
~> var table = [[6 10 2] [-2 0 10]]
~> put $table[0][1]
▶ 10
```

You can have a map where each value is another map, for example to represent
information about different entities (in this case, great ancient people)

```elvish-transcript Terminal - elvish
~> var people = [&Julius=[&title=Dictator &country=Rome]
                 &Alexander=[&title=King &country=Macedon]]
~> put $people[Julius][title]
▶ Dictator
```

You can even use lists and maps as map keys:

```elvish-transcript Terminal - elvish
~> var map-of-complex-keys = [&[&foo=bar]=map &[foo bar]=list]
~> put $map-of-complex-keys[[&foo=bar]]
▶ map
~> put $map-of-complex-keys[[foo bar]]
▶ list
```

(This example can be a bit of a brain teaser because all three meanings of `[`
and `]` are used. Using maps and lists as map keys is not super common, but it's
convenient when you do need it.)

The possibilities are limitless -- as long as the data fits in your computer's
RAM and the nesting relationship in your brain. And it's not just for
theoretical interest; for a real-world example of a list of maps with list
values, see
[the `update-servers-in-parallel.elv` case study](https://xiaq.me/draft.elv.sh/learn/scripting-case-studies.html#update-servers-in-parallel.elv).
(We'll learn more about some features in that example in
[Organizing and reusing code](organizing-and-reusing-code.html)).

# Strings

String is a value type that we have actually been using the whole time. Any text
not using any special punctuation or whitespaces are strings, as are quoted
strings. Unquoted strings are also known as **barewords**.

# Numbers

In fact, even numbers like `1` in Elvish are strings. Let's examine this
example:

```elvish-transcript Terminal - elvish
~> + 1 2
▶ (num 3)
~> + '1' '2'
▶ (num 3)
```

In this example, both `1` and `2` are strings, so `+ 1 2` and `+ '1' '2'` are
equivalent. Numerical commands like `+` accept strings and know how to treat
them as numbers internally.

This is OK for the most part, but there are situations where using strings as
numbers doesn't do what you need:

```elvish-transcript Terminal - elvish
~> put 1 | to-json
"1"
~> put 1 2 12 | order
▶ 1
▶ 12
▶ 2
```

(The [`order`](../ref/builtin.html#order) command reads value inputs, and
outputs them sorted.)

Elvish also supports a number type, and number values can be constructed using
the [`num`](../ref/builtin.html#num) command. You can use them as arguments to
numerical commands too, and they behave as numbers when converting to JSON or
sorting:

```elvish-transcript Terminal - elvish
~> num 3
▶ (num 3)
~> + (num 1) (num 2)
▶ (num 3)
~> num 1 | to-json
1
~> put (num 1) (num 2) (num 12) | order
▶ (num 1)
▶ (num 2)
▶ (num 12)
```

Since commands like `+` accept both `1` and `(num 1)`, we'll call both
"numbers". To distinguish the latter from the former, we usually call them
**typed numbers**.

As you have probably inferred from how the outputs were shown, even though
numerical commands accept both strings and typed numbers as arguments, they
always output typed numbers. This makes the result easier to use in contexts
like converting to JSON or sorting.

We have only worked with integers so far, but Elvish also supports rational
numbers and floating-point numbers:

```elvish-transcript Terminal - elvish
~> + 1/2 1/3
▶ (num 5/6)
~> + 0.2 0.3
▶ (num 0.5)
```

# Booleans

The [Boolean type](https://en.wikipedia.org/wiki/Boolean_data_type) has two
values, *true* and *false*, written in Elvish as two variables `$true` and
`$false`. Unsurprisingly, Elvish commands that tell you if something is true or
false output Boolean values:

```elvish-transcript Terminal - elvish
~> use str
~> str:has-suffix a.png .png
▶ $true
~> str:has-suffix a.png .jpg
▶ $false
```

Elvish also supports
[Boolean algebra](https://en.wikipedia.org/wiki/Boolean_algebra) operations:

```elvish-transcript Terminal - elvish
~> and $true $false
▶ $false
~> or $true $false
▶ $true
~> not $true
▶ $false
```

## Conditionals

Boolean values can be used to decide whether to do something or not, with the
help of the [`if`](../ref/language.html#if) command:

```elvish-transcript Terminal - elvish
~> if $true {
     echo "Yes it's true"
   }
Yes it's true
~> if $false {
     echo "This shouldn't happen"
   }
```

The `if` command is a
[**conditional**](https://en.wikipedia.org/wiki/Conditional_(computer_programming))
command, one of the most basic
[**control flows**](https://en.wikipedia.org/wiki/Control_flow). For loops,
which we have seen earlier, are another type of control flow.

Extending our previous example of converting JPG files to AVIF, let's add an
additional condition: we should only perform the conversion when the AVIF file
doesn't exist yet:

```elvish-transcript Terminal - elvish
~> use os
~> use str
~> for jpg [*.jpg] {
     var avif = (str:trim-suffix $jpg .jpg).avif
     if (not (os:exists $avif)) { # new condition
       gm convert $jpg $avif
     }
   }
```

# Value pipeline redux

Using what we've learned about values in Elvish, let's build a more interesting
value pipeline:

```elvish-transcript Terminal - elvish
~> curl -s https://xkcd.com/info.0.json | var latest = (from-json)[num] # ①
~> range $latest (- $latest 5) |                                        # ②
     each {|n| curl -s https://xkcd.com/$n/info.0.json } |              # ③
     from-json |                                                        # ④
     each {|info| echo $info[num]': '$info[title] }                     # ⑤
2905: Supergroup
2904: Physics vs. Magic
2903: Earth/Venus Venn Diagram
2902: Ice Core
2901: Geographic Qualifiers
```

The code above prints the latest 5 webcomics from <https://xkcd.com>, using its
[JSON API](https://xkcd.com/json.html). Let's examine it step by step:

1.  We first request <https://xkcd.com/info.0.json>, which fetches the
    information for the latest comic. We convert that to an Elvish map, and save
    the value of `num` in `$latest`.

2.  The [`range`](../ref/builtin.html#range) command outputs a range of numbers.
    The range can be increasing or decreasing, but in both cases, it starts at
    the first argument, and stops *before* it reaches the second argument.

    In the example output, `$latest` happens to be 2905, so `(- $latest 5)` is
    2900. You can see the `range` command in action by running it separately:

    ```elvish-transcript Terminal - elvish
    ~> range 2905 2900
    ▶ (num 2905)
    ▶ (num 2904)
    ▶ (num 2903)
    ▶ (num 2902)
    ▶ (num 2901)
    ```

3.  The [`each`](../ref/builtin.html#each) command runs the code inside `{` and
    `}`, assigning `$n` to each input (we will cover the syntax soon in
    [Organizing and reusing code](organizing-and-reusing-code.html)). It's
    similar to the `for` command we have seen before, except that it uses input
    values rather than list elements. The overall effect is the same as:

    ```elvish-transcript Terminal - elvish
    ~> curl -s https://xkcd.com/2095/info.0.json
       curl -s https://xkcd.com/2094/info.0.json
       curl -s https://xkcd.com/2093/info.0.json
       curl -s https://xkcd.com/2092/info.0.json
       curl -s https://xkcd.com/2091/info.0.json
    ...
    ```

    (To input multiple lines of commands at the prompt, press
    <kbd>Alt-Enter</kbd> instead of <kbd>Enter</kbd>.)

4.  The `from-json` command converts the stream of JSON outputs generated by all
    the `curl` commands into a stream of Elvish values, in this case maps.

5.  This second `each` command retrieves the values corresponding to the `num`
    and `title` keys, and print them.

## Limitations of value pipelines

Value pipelines allow you to manipulate data in a natural way similar to
traditional byte pipelines, but it does have an important shortcoming: it is not
available to external commands. We have already seen that external commands
can't produce value outputs; they also can't consume value inputs. As a result,
if you write an Elvish command that produces value outputs, you have to convert
it to bytes before an external command can make use of it, such as with the
`to-json` command.

# Conclusion

Elvish has a rich system of value types. These types allow you to model
real-world problems however you want, and consume and manipulate data sourced
from elsewhere. Elvish's value pipeline mechanism allows you to express these
operations in a fluid way.

Let's now move on to the final part of this series,
[Organizing and reusing code](organizing-and-reusing-code.html).
