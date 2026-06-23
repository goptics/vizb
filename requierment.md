## Make the group is optional for csv and json parsers

so I am thinking to introduce autogrouping feature which will active for only csv and jsons when user doesn't provide any sort of group value

the auto grouping logic will be simple.

first it find out the mnost unique column from the input and make it series or xAxis
and the rest numerical columns will become the yaxis. since its 2d data people can also apply --3d flags on it for eligble charts


no need to change any UI layer
