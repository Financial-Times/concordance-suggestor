Public API for Suggesting concordances
=========================================

Trying to formulate potential concordances between V1 and V2 organisations from the relationships to content


# Approach
To find a possible concordance look at the content that is annotated by the v1 uuid and the MAJOR_MENTIONS and then
look at the V2 MENTIONS relationships against the same content. The V2 org that is most MENTIONed against the same
content should be a contender as a concordance


#Cypher helpers

## Finding cordance suggestions
```
MATCH (organisation:Organisation{uuid:"a8279894-f09b-3c1c-b68e-143272afb82d"})<-[:MAJOR_MENTIONS]-(content:Content)
MATCH (content)-[v:MENTIONS]-(t:Thing) 
WITH t.uuid as uuid, t.prefLabel as prefLabel, count(t.uuid) as cnt
RETURN uuid,  prefLabel, cnt  ORDER BY cnt desc
```

## Used this to try and find possible unconcorded orgs
However we actually loads of other 'weird' annotations too

```MATCH (org:Organisation)<-[:MAJOR_MENTIONS]-(c:Content)
MATCH (org)<-[:IDENTIFIES]-(tme:TMEIdentifier)
OPTIONAL MATCH (org)<-[:IDENTIFIES]-(upp:UPPIdentifier)
OPTIONAL MATCH  (org)<-[rel:MENTIONS]-(c:Content)
WITH org, c, count(rel) as elCnt, count(upp) as uppCnt
WHERE elCnt=0 AND uppCnt=1
RETURN org, c
```

# Examples of unconcorded orgs
`http://spyglass.ft.com/organisations/c226fbf1-439b-3dd0-9c71-a7095950349e` 
Algorithm partially fails as there is not enough historic data so we get three suggestions of equal weight

`http://spyglass.ft.com/view/28bc5bf0-ab72-11e6-9cb3-bb8207902122?env=prod` 
Alphabet should be suggested but it is way down the list because often Editorial tag with Alphabet even if it is not mentioned at all in the article

Independent Press Standards Organisation
`http://spyglass.ft.com/organisations/6e4a6dad-2b29-350e-862e-e8a4f3d3a86b?env=prod` (V1 organisation here)
Comes up in a "MAJOR_MENTIONS" way in `http://spyglass.ft.com/view/feed23a0-a5b2-11e6-8b69-02899e8bd9d1?env=prod`
However no "MENTIONS" but it looks like there is a concordance between a V2 and another V1 Organisation (`TnN0ZWluX09OX0FGVE1fT05fMjU5NTQ0-T04=`)
Erroneous?

Slack Technologies, Inc
Looks like an unconcorded organisation:
V1 Uuid: `81db5e79-5709-330e-9d48-a85297b2c3b1`
V2 Uuid: `8aff8ea0-0c7e-31fb-8bd8-ac9178410f3a`
However the CES doesn't surface annotations because of a lack of the label of just "slack"
Example content: `http://spyglass.ft.com/view/50672400-a285-11e6-82c3-4351ce86813f?env=prod`
The algorithm failed again and even suggested Financial Times Ltd. Think this is again due to the lack of historical v1 annotations
Surprised the CES didn't suggest 
"Slack & Co" `1d2ec8b3-91f2-3e97-ac56-1a2d75a7a985` do have the label 'Slack'

Trafigura Ltd
`http://spyglass.ft.com/view/c8b626c5-dfc7-36e5-87f8-6fae778174f1`
V1: a8279894-f09b-3c1c-b68e-143272afb82d
V2: a2d51456-735f-3da5-9cfe-fd7f50f7f262
Algorithm works for this PERFECTLY :)

In the same article as for Trafigura Ltd. there is "PUMA" that also could be concorded however the algorithm suggested "Trafigura Beheer BV" also but as there is only 4 suggestions and three are obviously wrong then potentially it did work

Sky PLC versus Sky Plc - Where we have more than one major mentions!
`http://spyglass.ft.com/view/c34654d8-c066-11e6-81c2-f57d90f6741a`
