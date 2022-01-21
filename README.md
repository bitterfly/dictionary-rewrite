# Brief

This project is for the "Applications of Finite Automata" course at FMI (Faculty of Mathematics and Informatics) written in Go.

It builds a failure transducer which, given a dictionary D, represents Σ*D. Then it implements text rewriting using the leftmost longest strategy, i.e., given a text, it rewrites every occurence of a dictionary word in the text with its image.

The implementation follows the description in: https://www.aclweb.org/anthology/W17-4001.pdf

The rewrite is done in O(n) where n is the length of the text.

# Usage:

## Requirements

An installed version of Go (at least 1.14).

Mind that large dictionaries take a lot of RAM (a.k.a. 25GB for 1.4GB dictionary ~ 18,000,000 words)

## Instalation

    $ go install github.com/bitterfly/dictionary-rewrite@latest


## Running

    $ dictionary-rewrite <dictionary-file> <text-file>
    
where `<dictionary-file>` is in the format

    <non-empty-string>\t<string>

and is sorted

and `<text-file>` is any UTF-8 text file

## Example
```
> dictionary-rewrite example-dictionary.txt example-text.txt
```
yields

```
Имало някога едно малко сладко непълнолетно мотористче. Всеки го обиквал от пръв поглед, но най-много го обичала старата му мотористка, която всеки път се чудела какво да даде на детето. Веднъж му подарила каска от червено кадифе, която му стояла тъй хубаво, че то не искало да носи друга и затова хората почнали да го наричат Кървавата каска
...
```