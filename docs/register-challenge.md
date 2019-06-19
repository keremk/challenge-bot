# Register/Edit a Challenge

## Register the Challenge Template
Before sending a new challenge, you need to register the challenge template into the database. To do that you will need to type in the below: (We recommend you to create a specific Slack channel for all the coding challenges you will be sending. If there are multiple challenges for different positions, it is better to create different channels for each in Slack)

* Go to your Slack channel and type:

```
  /challenge new
```

* You will be presented by the below dialog:

![Challenge New Dialog](slack-new-challenge.png)

* In the above dialog:
  * *Challenge Name* Give a user friendly and unique name to the challenge, you will be using this later to pick a template for the challenge.
  * *Template Repo Name* The name of the Github Challenge Template Repo that you have registered before. E.g. challenge_temp1 from the [previously created challenge repo in Github](github-workflow.md)
  * *Repo Name Format* You can specify a naming format for repos that the tool will be creating for the candidates. The default is already populated so you can either keep it or modify it. Use `CHALLENGENAME` as a placeholder for the name of the challenge and `GITHUBALIAS` as a placeholder for the candidate's github alias.
  * *Github Account Name* Specify the Github account the challenge repos (and their templates) will be (are) stored.  
* And once you are comfortable tap `Create` button. This will register the coding challenge template.

## Edit the challenge template
You can also edit the challenge template you created.

* Go to your Slack channel and type:

```
  /challenge edit CHALLENGENAME
```

* You will be presented by the below dialog:

![Challenge New Dialog](slack-new-challenge.png)

And edit the same was as in new registration.

