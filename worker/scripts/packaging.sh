printf "Artifact packaging\n";
if [ -z "$ARTIFACT_DIR" ]
then
  # Артефактом является весь проект целиком, архивируем всю папку проекта
  printf "Packaging the whole project start\n";
  ls -1 builds/project-$ID/;
  zip -rq $ARTIFACT_ZIP_NAME builds/project-$ID/*;
  printf "Packaging the whole project end\n";
else
  # Артефактом является папка или файл проекта
  if [[ -d $ARTIFACT_DIR ]]; then
    printf "Packaging only \"$ARTIFACT_DIR\" dir\n";
  elif [[ -f $ARTIFACT_DIR ]]; then
    printf "Packaging only \"$ARTIFACT_DIR\" file\n";
  else
    printf "Nothing to packaging, artifact \"$ARTIFACT_DIR\" not found\n";
    exit 1
  fi

  # Архивируем если артефакт найден
  ls -1 builds/project-$ID/$ARTIFACT_DIR/;
  zip -r $ARTIFACT_ZIP_NAME builds/project-$ID/$ARTIFACT_DIR/*;
  printf "Packaging success finished\n";
fi
exit 0;