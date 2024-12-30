package translationflow

const (
	OneChunkInitialTranslationSystemMessageTmpl = `
You are an expert linguist, specializing in translation from {{.SourceLang}} to {{.TargetLang}}.
`
	OneChunkInitialTranslationPromptTmpl = `
This is an {{.SourceLang}} to {{.TargetLang}} translation, please provide the {{.TargetLang}} translation for this text. \
Do not provide any explanations or text apart from the translation.
{{.SourceLang}}: {{.SourceText}}

{{.TargetLang}}:
`
	OneChunkReflectOnTranslationSystemMessageTmpl = `
You are an expert linguist specializing in translation from {{.SourceLang}} to {{.TargetLang}}. \
You will be provided with a source text and its translation and your goal is to improve the translation.
`
	OneChunkReflectOnTranslationWithCountryPromptTmpl = `
Your task is to carefully read a source text and a translation from {{.SourceLang}} to {{.TargetLang}}, and then give constructive criticism and helpful suggestions to improve the translation. \
The final style and tone of the translation should match the style of {{.TargetLang}} colloquially spoken in {{.Country}}.

The source text and initial translation, delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT> and <TRANSLATION></TRANSLATION>, are as follows:

<SOURCE_TEXT>
{{.SourceText}}
</SOURCE_TEXT>

<TRANSLATION>
{{.Translation1}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's \n\
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n\
(ii) fluency (by applying {{.TargetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),\n\
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),\n\
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.TargetLang}}).\n\

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.`

	OneChunkReflectOnTranslationPromptTmpl = `
Your task is to carefully read a source text and a translation from {{.SourceLang}} to {{.TargetLang}}, and then give constructive criticisms and helpful suggestions to improve the translation. \

The source text and initial translation, delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT> and <TRANSLATION></TRANSLATION>, are as follows:

<SOURCE_TEXT>
{{.SourceText}}
</SOURCE_TEXT>

<TRANSLATION>
{{.Translation1}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's \n\
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n\
(ii) fluency (by applying {{.TargetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),\n\
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),\n\
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.TargetLang}}).\n\

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.
`
	OneChunkImprovementTranslationSystemMessageTmpl = `
You are an expert linguist, specializing in translation editing from {{.SourceLang}} to {{.TargetLang}}.
`
	OneChunkImprovementTranslationPromptTmpl = `
Your task is to carefully read, then edit, a translation from {{.SourceLang}} to {{.TargetLang}}, taking into
account a list of expert suggestions and constructive criticisms.

The source text, the initial translation, and the expert linguist suggestions are delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT>, <TRANSLATION></TRANSLATION> and <EXPERT_SUGGESTIONS></EXPERT_SUGGESTIONS> \
as follows:

<SOURCE_TEXT>
{{.SourceText}}
</SOURCE_TEXT>

<TRANSLATION>
{{.Translation1}}
</TRANSLATION>

<EXPERT_SUGGESTIONS>
{{.Reflection}}
</EXPERT_SUGGESTIONS>

Please take into account the expert suggestions when editing the translation. Edit the translation by ensuring:

(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {{.TargetLang}} grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions), \
(iii) style (by ensuring the translations reflect the style of the source text)
(iv) terminology (inappropriate for context, inconsistent use), or
(v) other errors.

Output only the new translation and nothing else.
`
)

const (
	MultiChunkTranslationSystemMessageTmpl = "You are an expert linguist, specializing in translation from {{.SourceLang}} to {{.TargetLang}}."

	MultiChunkTranslationPromptTmpl = `Your task is to provide a professional translation from {{.SourceLang}} to {{.TargetLang}} of PART of a text.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>. Translate only the part within the source text
delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS>. You can use the rest of the source text as context, but do not translate any
of the other text. Do not output anything other than the translation of the indicated part of the text.

<SOURCE_TEXT>
{{.TaggedText}}
</SOURCE_TEXT>

To reiterate, you should translate only this part of the text, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.ChunkToTranslate}}
</TRANSLATE_THIS>

Output only the translation of the portion you are asked to translate, and nothing else.`

	MultiChunkReflectionSystemMessageTmpl = "You are an expert linguist specializing in translation from {{.SourceLang}} to {{.TargetLang}}. You will be provided with a source text and its translation and your goal is to improve the translation."

	MultiChunkReflectionWithCountryPromptTmpl = `Your task is to carefully read a source text and part of a translation of that text from {{.SourceLang}} to {{.TargetLang}}, and then give constructive criticism and helpful suggestions for improving the translation.
The final style and tone of the translation should match the style of {{.TargetLang}} colloquially spoken in {{.Country}}.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
as context for critiquing the translated part.

<SOURCE_TEXT>
{{.TaggedText}}
</SOURCE_TEXT>

To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.ChunkToTranslate}}
</TRANSLATE_THIS>

The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
<TRANSLATION>
{{.TranslationChunk}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's:
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {{.TargetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.TargetLang}}).

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.`

	MultiChunkReflectionPromptTmpl = `Your task is to carefully read a source text and part of a translation of that text from {{.SourceLang}} to {{.TargetLang}}, and then give constructive criticism and helpful suggestions for improving the translation.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
as context for critiquing the translated part.

<SOURCE_TEXT>
{{.TaggedText}}
</SOURCE_TEXT>

To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.ChunkToTranslate}}
</TRANSLATE_THIS>

The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
<TRANSLATION>
{{.TranslationChunk}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's:
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {{.TargetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.TargetLang}}).

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.`

	MultiChunkImprovementSystemMessageTmpl = "You are an expert linguist, specializing in translation editing from {{.SourceLang}} to {{.TargetLang}}."

	MultiChunkImprovementPromptTmpl = `Your task is to carefully read, then improve, a translation from {{.SourceLang}} to {{.TargetLang}}, taking into
account a set of expert suggestions and constructive criticisms. Below, the source text, initial translation, and expert suggestions are provided.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
as context, but need to provide a translation only of the part indicated by <TRANSLATE_THIS> and </TRANSLATE_THIS>.

<SOURCE_TEXT>
{{.TaggedText}}
</SOURCE_TEXT>

To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.ChunkToTranslate}}
</TRANSLATE_THIS>

The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
<TRANSLATION>
{{.TranslationChunk}}
</TRANSLATION>

The expert translations of the indicated part, delimited below by <EXPERT_SUGGESTIONS> and </EXPERT_SUGGESTIONS>, are as follows:
<EXPERT_SUGGESTIONS>
{{.ReflectionChunk}}
</EXPERT_SUGGESTIONS>

Taking into account the expert suggestions rewrite the translation to improve it, paying attention
to whether there are ways to improve the translation's

(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {{.TargetLang}} grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions),
(iii) style (by ensuring the translations reflect the style of the source text)
(iv) terminology (inappropriate for context, inconsistent use), or
(v) other errors.

Output only the new translation of the indicated part and nothing else.`
)
