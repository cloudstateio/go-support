lazy val docs = project
  .in(file("."))
  .enablePlugins(CloudstateParadoxPlugin)
  .settings(
    deployModule := "go",
    paradoxProperties in Compile ++= Map(
      "cloudstate.go.version" -> "1.14",
      "cloudstate.go.lib.version" -> { if (isSnapshot.value) previousStableVersion.value.getOrElse("0.0.0") else version.value },
      "extref.cloudstate.base_url" -> "https://cloudstate.io/docs/core/current/%s"
    )
  )
