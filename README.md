# Brief

This project is for the "Applications of Finite Automata" course at FMI (Faculty of Mathematics and Informatics) written in Go.

It builds a failure transducer which, given a dictionary D, represents Σ*D. Then it implements text rewriting using the leftmost long strategy, i.e., given a text, it rewrites every occurence of a dictionary word in the text with its image.

The implementation follows the description in: https://www.aclweb.org/anthology/W17-4001.pdf

# Usage:

## Requirements

An installed version of Go (at least 1.14).

Mind that large dictionaries take a lot of RAM (a.k.a. 25GB for 1.4GB dictionary ~ 18,000,000 words)

## Instalation

    $ go install github.com/bitterfly/ftransducer/ftransducer_cmd


## Running

    $ ftransducer_cmd <dictionary-file> <text-file>
    
where <dictionary-file> is in the format

    <non-empty-string>\t<string>

and <text-file> is any UTF-8 text file
