printf "Adding a Private Key\n";
eval `ssh-agent -s`;
ssh-add - <<< "${SSH_KEY}";
if [ $? -eq 0 ]; then
    printf "Key added successfully\n";
else
    printf "An error occurred while adding the key\n";
    exit 1;
fi
git clone git@github.com:$NAME.git builds/project-$ID;
if [ $? -eq 0 ]; then
    printf "The project is successfully inclined\n";
else
    printf "An error occurred while cloning the project\n";
    exit 1;
fi
exit 0;