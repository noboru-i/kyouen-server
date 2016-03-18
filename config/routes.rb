Rails.application.routes.draw do
  root 'welcome#index'

  get '/kyouens', to: 'kyouens#index'
end
