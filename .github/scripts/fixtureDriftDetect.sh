#!/usr/bin/env bash

set -euo pipefail

showHelp() {
	cat << EOF

Usage: ./fixtureDriftDetect.sh (file|dir|prefix) [...] [--revision <rev>] [--regex <expr>]

Positional arguments:
    One or more files, directories, or prefixes to process.

Options:
    -h, -help, --help       Display help.
    -revision, --revision   Specify the revision to compare with.
                            Default: origin/main
    -regex, --regex         Specify regular expression for lines to compare.
                            Default: '(id: |url: |status: )'

Dependencies:
    - diff
    - git
    - grep

Examples:
    ./fixtureDriftDetect.sh scc/provider/fixtures/
    ./fixtureDriftDetect.sh scc/provider/resources/fixtures/
    ./fixtureDriftDetect.sh \
        scc/provider/datasources/fixtures/ \
        scc/provider/listresources/fixtures/ \
        scc/provider/resources/fixtures/

EOF
}

revision='origin/main'
regex='(id: |url: |status: )'

pathspecs=()

while [[ $# -gt 0 ]]; do
	case "$1" in
		-h|--help)
			showHelp
			exit 0
			;;

		-revision|--revision)
			if [[ $# -lt 2 ]]; then
				>&2 echo "missing value for $1"
				exit 1
			fi

			revision="$2"
			shift 2
			;;

		-regex|--regex)
			if [[ $# -lt 2 ]]; then
				>&2 echo "missing value for $1"
				exit 1
			fi

			regex="$2"
			shift 2
			;;

		--*|-*)
			>&2 echo "unknown option $1"
			exit 1
			;;

		*)
			pathspecs+=("$1")
			shift
			;;
	esac
done

if [[ ${#pathspecs[@]} -eq 0 ]]; then
	>&2 echo "missing positional argument (file|dir|prefix)"
	exit 1
fi

found=false
failed=false

for pathspec in "${pathspecs[@]}"; do

	matched=false

	for file in ${pathspec}*; do

		if [[ -f "$file" ]]; then
			found=true
			matched=true

			if ! git cat-file -e "$revision:$file" 2>/dev/null; then
				printf "ignoring file '%s' since it does not exist in revision '%s'\n" \
					"$file" \
					"$revision"
				continue
			fi

			diff_output=$(
				diff \
					<(git cat-file blob "$revision:$file" | grep -E "$regex" || true) \
					<(grep -E "$regex" "$file" || true) \
					|| true
			)

			if [[ -n "$diff_output" ]]; then
				printf "\nChanges in file '%s' to revision '%s' with regards to regex '%s':\n\n%s\n\n" \
					"$file" \
					"$revision" \
					"$regex" \
					"$diff_output"

				failed=true
			fi
		fi

	done

	if [[ "$matched" == false ]]; then
		printf "WARNING: no files matched pathspec '%s'\n" "$pathspec"
	fi

done

if [[ "$found" == false ]]; then
	printf "\nERROR: no files processed\n\n"
	exit 1
fi

if [[ "$failed" == true ]]; then
	printf "\nWARNING: Changes to revision '%s' detected with regards to regex '%s'\n\n" \
		"$revision" \
		"$regex"

	exit 1
fi

printf "\nSUCCESS: No changes to revision '%s' detected with regards to regex '%s'\n\n" \
	"$revision" \
	"$regex"

exit 0