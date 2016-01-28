Say hi to your friends through Bonjour by creating (empty) Remote Shares in
their Finder.

Install
-------

1. Install Go by [downloading a package](https://golang.org/dl/) or on OS X with [Homebrew](http://brew.sh/) like `brew install go` (or for the beta, `brew install --devel go`). You might have to open a new shell to get the env variables you need.
2. Get bonjourno: `go get -v github.com/subparlabs/bonjourno`
3. Try running `bonjourno`. If it didn't find it, you probably don't have `$GOPATH/bin` in your path, so just run it from there: `$GOPATH/bin/bonjourno`


Running
-------

Bonjourno has several ways of combining where to get data, how to interpret it, and how to go through it.

### Data Source
1. Just say something straight in the command line: `bonjourno This will show up in Finder`
2. Point it to a file: `bonjourno --file=messages.txt`
3. Point it to a url: `bonjourno --url='https://raw.githubusercontent.com/SubparLabs/bonjourno/master/README.md'`

### Data Format
1. The default is to make every line a message, just keeping the first 40 characters.
2. Instead of lines, go through all the text in the data, in groups of words < 20 characters: `bonjourno --file=essay.txt --words`
3. Grab fields from CSV data by specified the index of the field: `bonjourno --file=data.csv --csv-field=2`

### Iteration
1. By default, it will sequentially go through the messages in order.
2. You can randomize the order: `bonjourno --file=messages.txt --random`

### Misc
1. You can prefix messages, for ex to keep them at the top of the list: `bonjourno --file=messages.txt --prefix='1-'`
2. Specify how frequently messages should change: `bonjourno --file=messages.txt --interval=10s`


Examples
--------

1. Random countries: `bonjourno --interval=10s --random --csv-field=1 --url='https://raw.githubusercontent.com/datasets/un-locode/master/data/country-codes.csv'`
2. Companies names from Crunchbase: `bonjourno --interval=10s --random --csv-field=0 --url='https://raw.githubusercontent.com/datasets/crunchcrawl/master/companydata.csv'`
3. Companies listed in the New York Stock Exchange: `bonjourno --interval=10s --random --csv-field=1 --url='https://raw.githubusercontent.com/datasets/nyse-listings/master/data/nyse-listed.csv'`

Run with `--help` for usage. If you have problems, open an [Issue](https://github.com/SubparLabs/bonjourno/issues).
