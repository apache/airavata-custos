package org.apache.custos.sharing.service.core.models;

public class SearchCriteria {
    private EntitySearchField searchField;
    private String value;
    private SearchCondition searchCondition;

    public EntitySearchField getSearchField() {
        return searchField;
    }

    public void setSearchField(EntitySearchField searchField) {
        this.searchField = searchField;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public SearchCondition getSearchCondition() {
        return searchCondition;
    }

    public void setSearchCondition(SearchCondition searchCondition) {
        this.searchCondition = searchCondition;
    }
}
