printf "Deploy project\n";
if [ -z "$ARTIFACT_ZIP" ]
then
  printf "Nothing to deploy\n";
  exit 1;
else
  unzip -o $ARTIFACT_ZIP -d $DEPLOY_DIR
fi
exit 0;