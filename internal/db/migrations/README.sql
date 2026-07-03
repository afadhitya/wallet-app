Migration files use the pattern `NNN_description.sql` where NNN is a zero-padded
integer version number. Migrations are applied in alphabetical order based on
file name. Only files matching the version prefix pattern are executed.

