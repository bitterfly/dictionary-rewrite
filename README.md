# Brief

This project is for applications of automata  course at FMI(Faculty of Mathematics and Informatics) written in GO.

It builds a failure transducer which, given a dictionary D, represents Σ*D. Then it implements text rewriting using the leftmost long strategy, i.e.
given a text it rewrites every occurence of a dictionary word in the texts with its image.

The implementation follows the description in: https://www.aclweb.org/anthology/W17-4001.pdf

# Usage:

## Requirements

An installed version of Go.
Mind that large dictionaries take a lot of RAM (a.k.a. 25GB for 1.4GB dictionary ~ 18,000,000 words)

## Instalation

    $ go install go install github.com/bitterfly/ftransducer/ftransducer_cmd


## Running

    $ ftransducer_cmd <dictionary-file> <text-file>
    
where <dictionary-file> is in the format

<non-empty-string>\t<string>

and <text-file> is any UTF-8 text file
